package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/landonturner/scheduler/internal/api"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func main() {
	db, err := gorm.Open("sqlite3", "sqlite.db")
	if err != nil {
		log.Fatal(err)
	}

	var jwtSecret []byte
	jwtSecretString := os.Getenv("JWT_SECRET")
	if jwtSecretString == "" {
		log.Println("WARNING: Use the env var JWT_SECRET (32 bytes in hex) for persistent sessions")
		jwtSecret = make([]byte, 32)
		rand.Read(jwtSecret)
		log.Println("Using jwt secret: ", hex.EncodeToString(jwtSecret))
	} else {
		jwtSecret, err = hex.DecodeString(jwtSecretString)
		if err != nil {
			log.Fatal("Error decoding jwt secret. Provide a 32 byte hex string")
		}
	}

	url := os.Getenv("REMOTE_URL")
	if url == "" {
		url = "http://localhost:1337/status"
		log.Println("WARNING: Using dummy web endpoint url. Set REMOTE_URL env var.")
	}

	routes := api.NewRoutes(db, jwtSecret, api.NewHTTPClient(url))
	routes.MigrateDB()

	r := mux.NewRouter()

	a := r.PathPrefix("/").Subrouter()
	a.Use(routes.AuthMiddleware)

	// All api requests must be authenticated
	a.HandleFunc("/me", routes.Me).Methods("GET")
	a.HandleFunc("/schedules", routes.ListSchedules).Methods("GET")
	a.HandleFunc("/schedules", routes.CreateSchedule).Methods("POST")
	a.HandleFunc("/schedules/{id}", routes.DeleteSchedule).Methods("DELETE")

	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "We're up doc")
	}).Methods("GET")

	// Login should not be under the AuthMiddleware
	r.HandleFunc("/login", routes.LoginFunc).Methods("POST")
	r.HandleFunc("/register", routes.RegisterFunc).Methods("POST")

	if _, err := os.Stat("client/dist"); os.IsNotExist(err) {
		log.Println("Could not find client/dist/index.html Run client build please")
	} else {
		r.PathPrefix("/").Handler(fileServerWithIndexFallback())
	}

	logRoutes(r)

	log.Println("Checking schedules every 10 seconds for deploy to run")
	go func() {
		for {
			routes.CheckSchedules()
			time.Sleep(10 * time.Second)
		}
	}()

	loggedRoutes := handlers.LoggingHandler(os.Stdout, r)

	log.Printf("Scheduler server listening on 0.0.0.0:1337\n\n")
	log.Fatal(http.ListenAndServe("0.0.0.0:1337", loggedRoutes))
}

func logRoutes(router *mux.Router) {
	fmt.Println("--- Routes ---")
	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		methods, err := route.GetMethods()
		if err == nil {
			fmt.Printf(" %6s", strings.Join(methods, ","))
		} else {
			return nil
		}
		pathTemplate, err := route.GetPathTemplate()
		if err == nil {
			fmt.Printf(" %s\n", pathTemplate)
		}

		return nil
	})
	fmt.Println("--------------")
	fmt.Println()

	if err != nil {
		log.Fatal(err)
	}
}

func fileServerWithIndexFallback() http.Handler {
	fs := http.Dir("client/dist")
	fsh := http.FileServer(fs)
	dat, err := ioutil.ReadFile("client/dist/index.html")
	if err != nil {
		log.Fatalln("Error reading client/dist/index.html", err.Error())
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := fs.Open(path.Clean(r.URL.Path))
		if os.IsNotExist(err) {
			w.Write(dat)
			return
		}
		fsh.ServeHTTP(w, r)
	})
}
