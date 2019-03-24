package api

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestLoginFunc(t *testing.T) {
	// create dummy db
	f, _ := ioutil.TempFile("", "")
	db, err := gorm.Open("sqlite3", f.Name())
	defer os.Remove(f.Name())
	defer db.Close()

	if err != nil {
		t.Fatal("Error initializing test sqlite db: " + err.Error())
	}

	routes := NewRoutes(db, []byte{}, &HTTPClient{})
	routes.MigrateDB()

	password := "s3cure_pw"
	// create dummy user
	u := User{
		Email: "dummyuser",
		Hash:  createPasswordHash(password),
	}

	db.Create(&u)
	handler := http.HandlerFunc(routes.LoginFunc)

	// test correct login
	t.Run("correct login", func(t *testing.T) {
		payload := "email=" + u.Email + "&password=" + password
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(payload)))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		body, err := ioutil.ReadAll(rr.Body)
		if err != nil {
			t.Error("Error reading body", err.Error())
		}

		if status := rr.Code; status != http.StatusOK {
			t.Error(string(body))
			t.Errorf("Status was not 200 for correct login: %d\n", status)
		}

		email, err := routes.extractEmailFromJWT(string(body))
		if err != nil {
			t.Error(string(body))
			t.Fatal("Error decoding jwt", err.Error())
		}

		if email != u.Email {
			t.Error("User email not encoded in jwt properly")
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		payload := "email=" + u.Email + "&password=" + password + "buttz"
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(payload)))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Status was not 401 for incorrect login: %d\n", status)
		}
	})

	t.Run("wrong email", func(t *testing.T) {
		payload := "email=" + u.Email + "stuff" + "&password=" + password
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(payload)))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Status was not 401 for incorrect login: %d\n", status)
		}
	})

	t.Run("missing email", func(t *testing.T) {
		payload := "password=" + password
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(payload)))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Status was not 400 for no email: %d\n", status)
		}
	})

	t.Run("missing password", func(t *testing.T) {
		payload := "email=" + u.Email
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(payload)))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Status was not 400 for no password: %d\n", status)
		}
	})

	t.Run("malformed body", func(t *testing.T) {
		payload := "all work and no play makes nate a dull boy"
		req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer([]byte(payload)))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Status was not 400 for garbage: %d\n", status)
		}
	})

	t.Run("no body", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/login", nil)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Status was not 400 for no body: %d\n", status)
		}
	})
}

func TestRegisterFunc(t *testing.T) {
	// create dummy db
	f, _ := ioutil.TempFile("", "")
	db, err := gorm.Open("sqlite3", f.Name())
	defer os.Remove(f.Name())
	defer db.Close()

	if err != nil {
		t.Fatal("Error initializing test sqlite db")
	}

	routes := NewRoutes(db, []byte{}, &HTTPClient{})
	routes.MigrateDB()
	handler := http.HandlerFunc(routes.RegisterFunc)

	email := "me@example.com"
	password := "d0pep4ssw0rd"
	name := "dudeman"
	validUserPayload := "email=" + email + "&password=" + password + "&name=" + name

	testHarness := []struct {
		testName string
		payload  string
		status   int
	}{
		{testName: "no body", payload: "", status: http.StatusBadRequest},
		{testName: "malformed body", payload: "this isn't even!", status: http.StatusBadRequest},
		{testName: "no password", payload: "email=" + email + "&name=" + name, status: http.StatusBadRequest},
		{testName: "no email", payload: "password=" + password + "&name=" + name, status: http.StatusBadRequest},
		{testName: "no name", payload: "email=" + email + "&password=" + password, status: http.StatusBadRequest},
		{testName: "vaild user", payload: validUserPayload, status: http.StatusCreated},
	}

	for _, th := range testHarness {
		t.Run(th.testName, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer([]byte(th.payload)))
			req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; th.status != status {
				t.Errorf("Incorrect status. Expected: %d, Actual: %d\n", th.status, status)
			}
		})
	}

	t.Run("valid user was created", func(t *testing.T) {
		u := User{}
		db.Where("email = ?", email).First(&u)
		if u.ID == 0 {
			t.Error("User was not created")
		}

		if u.Name != name {
			t.Errorf("Name was saved incorrectly. Exptected: %s Got: %s\n", name, u.Name)
		}

		if u.Email != email {
			t.Errorf("Email was not saved correctly. Expected: %s, Got: %s\n", email, u.Email)
		}

		if !verifyPassword(password, u.Hash) {
			t.Error("Password hashed incorrectly")
		}
	})

	t.Run("cannot create same user twice", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer([]byte(validUserPayload)))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; http.StatusBadRequest != status {
			t.Errorf("Incorrect status. Expected: %d, Actual: %d\n", http.StatusBadRequest, status)
		}
	})
}

func TestAuthMiddleware(t *testing.T) {
	// create dummy db
	f, _ := ioutil.TempFile("", "")
	db, err := gorm.Open("sqlite3", f.Name())
	defer os.Remove(f.Name())
	defer db.Close()

	if err != nil {
		t.Fatal("Error initializing test sqlite db")
	}

	routes := NewRoutes(db, []byte{}, &HTTPClient{})
	routes.MigrateDB()

	u := User{
		Email: "me@email.email",
	}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := r.Context().Value(emailContextKey)
		if i == nil {
			t.Error("Email was not present on context")
		} else {
			email := r.Context().Value(emailContextKey).(string)
			if email != u.Email {
				t.Errorf("Email not set on context correctly. Expected: %s, Got: %s\n", u.Email, email)
			}
			w.WriteHeader(http.StatusOK)
		}
	})
	handlerToTest := routes.AuthMiddleware(nextHandler)

	t.Run("no session", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()

		handlerToTest.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("Status was not 401 with no auth: %d\n", status)
		}
	})

	t.Run("session present", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/me", nil)
		rr := httptest.NewRecorder()

		jwt, _ := routes.createJWT(u)
		req.Header["Authorization"] = []string{fmt.Sprintf("Bearer %s", jwt)}

		handlerToTest.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Status was not 200 with jwt present: %d\n", status)
		}
	})
}

func TestJWT(t *testing.T) {

	routes := NewRoutes(nil, []byte{}, nil)

	user := User{
		Email: "billybob@thing.thing",
	}

	jwt, err := routes.createJWT(user)
	if err != nil {
		t.Error("Error creating jwt: ", err.Error())
	}

	if len(jwt) == 0 {
		t.Error("Jwt not created")
	}

	email, err := routes.extractEmailFromJWT(jwt)

	if err != nil {
		t.Error("Error extracting email", err.Error())
	}

	if email != user.Email {
		t.Error("Email not correct")
	}
}

func TestPasswordHashing(t *testing.T) {
	pw := "im the d0pest password"

	hash := createPasswordHash(pw)

	if len(hash) == 0 {
		t.Fatal("Password hash was not populated")
	}

	t.Run("correct password", func(t *testing.T) {
		if !verifyPassword(pw, hash) {
			t.Error("Password was rejected incorrectly")
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		if verifyPassword("not right", hash) {
			t.Error("Password was accepted incorrectly")
		}
	})

	t.Run("Malformed hash", func(t *testing.T) {
		if verifyPassword(pw, "garbagesalt") {
			t.Error("Garbage salt not rejected")
		}
	})
}
