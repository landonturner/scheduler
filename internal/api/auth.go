package api

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/pbkdf2"
)

type key int

var emailContextKey key = 1

// AuthMiddleware provides the http.Handler for authentication
func (routes *Routes) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header["Authorization"]
		if len(authHeader) == 0 {
			writeErrorMessage(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		authString := authHeader[0]

		authParts := strings.Split(authString, " ")
		if len(authParts) != 2 {
			writeErrorMessage(w, "Unauthorized", http.StatusUnauthorized)
			log.Println("Unexpected auth header:", authString)
			return
		}

		token := authParts[1]

		email, err := routes.extractEmailFromJWT(token)
		if err != nil {
			writeErrorMessage(w, "Unauthorized", http.StatusUnauthorized)
			log.Printf("Error parsing jwt %s error: %s\n", token, err.Error())
			return
		}

		// set the email back on the context
		c := context.WithValue(r.Context(), emailContextKey, email)
		r = r.WithContext(c)

		next.ServeHTTP(w, r)
	})
}

// LoginFunc handles logins and assigns session tokens
func (routes *Routes) LoginFunc(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if email == "" {
		writeErrorMessage(w, "Email required", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		writeErrorMessage(w, "Password required", http.StatusBadRequest)
		return
	}

	u := User{}
	routes.db.Where("email = ?", email).First(&u)

	if u.ID == 0 {
		writeErrorMessage(w, "Incorrect username or password", http.StatusUnauthorized)
		return
	}

	if verifyPassword(password, u.Hash) {
		// build a jwt and return it here
		jwt, err := routes.createJWT(u)
		if err != nil {
			writeErrorMessage(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(jwt))
	} else {
		writeErrorMessage(w, "Incorrect username or password", http.StatusUnauthorized)
	}
}

// RegisterFunc handles registrations and assigns session tokens
func (routes *Routes) RegisterFunc(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if email == "" {
		writeErrorMessage(w, "Email required", http.StatusBadRequest)
		return
	}

	password := r.FormValue("password")
	if password == "" {
		writeErrorMessage(w, "Password required", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	if name == "" {
		writeErrorMessage(w, "Name required", http.StatusBadRequest)
		return
	}

	u := User{}
	routes.db.Where("email = ?", email).First(&u)
	if u.ID != 0 {
		writeErrorMessage(w, "Email previously registered", http.StatusBadRequest)
		return
	}

	h := createPasswordHash(password)

	// might want to validate user here

	u = User{
		Email: email,
		Hash:  h,
		Name:  name,
	}

	routes.db.Create(&u)
	w.WriteHeader(http.StatusCreated)
}

func (routes *Routes) createJWT(user User) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": user.Email,
	})

	return token.SignedString(routes.jwtSecret)
}

func (routes *Routes) extractEmailFromJWT(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// validate the alg
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return routes.jwtSecret, nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if claims["email"] != nil {
			email := claims["email"].(string)
			return email, nil
		}
		return "", errors.New("Email not present on claims")
	}

	return "", errors.New("Error parsing claims")
}

var iterations = 10000
var keyLength = 32

func createPasswordHash(password string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	k := pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha1.New)
	return hex.EncodeToString(salt) + ":" + hex.EncodeToString(k)
}

func verifyPassword(password string, hash string) bool {
	p := strings.Split(hash, ":")
	salt, err := hex.DecodeString(p[0])
	if err != nil {
		fmt.Print("Error decoding salt from " + hash)
		return false
	}
	key := p[1]
	k := pbkdf2.Key([]byte(password), salt, iterations, keyLength, sha1.New)
	return key == hex.EncodeToString(k)
}
