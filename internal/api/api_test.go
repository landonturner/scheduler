package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func TestCreateSchedule(t *testing.T) {
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

	u := User{Email: "person@email.com", Name: "Dude Man"}
	routes.db.Create(&u)

	t.Run("no time given", func(t *testing.T) {
		payload := ""

		req, _ := http.NewRequest("POST", "/schedules", bytes.NewBuffer([]byte(payload)))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		c := context.WithValue(req.Context(), emailContextKey, u.Email)
		req = req.WithContext(c)

		http.HandlerFunc(routes.CreateSchedule).ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("Incorrect status, expected: 400, got: %d\n", status)
		}
	})

	t.Run("valid time given", func(t *testing.T) {
		tyme := time.Now()

		timeString := tyme.Format(time.RFC3339)

		payload := "time=" + timeString

		req, _ := http.NewRequest("POST", "/schedules", bytes.NewBuffer([]byte(payload)))
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		rr := httptest.NewRecorder()
		c := context.WithValue(req.Context(), emailContextKey, u.Email)
		req = req.WithContext(c)

		http.HandlerFunc(routes.CreateSchedule).ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("Incorrect status, expected: 201, got: %d\n", status)
		}

		sched := Schedule{}
		routes.db.First(&sched)

		if sched.ID == 0 {
			t.Error("Schedule not saved in db correctly")
		}
	})
}

func TestDeleteSchedule(t *testing.T) {
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

	s := Schedule{
		Source: "me",
		Time:   time.Now(),
	}

	db.Create(&s)

	path := fmt.Sprintf("/schedules/%d", s.ID)
	req, _ := http.NewRequest("DELETE", path, nil)
	c := context.WithValue(req.Context(), emailContextKey, "fake@email.com")
	req = req.WithContext(c)

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	router.HandleFunc("/schedules/{id}", routes.DeleteSchedule).Methods("DELETE")
	router.ServeHTTP(rr, req)

	db.Unscoped().First(&s)

	if s.DeletedAt == nil {
		t.Error("Did not delete the entry correctly")
	}
}

func TestListSchedules(t *testing.T) {
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

	u := User{Email: "person@email.com"}
	routes.db.Create(&u)

	sched := Schedule{
		Time:   time.Now(),
		Source: "billybob",
	}
	routes.db.Create(&sched)

	sched = Schedule{
		Time:   time.Now(),
		Source: "billybob",
	}
	routes.db.Create(&sched)

	req, _ := http.NewRequest("GET", "/schedules", nil)
	rr := httptest.NewRecorder()
	c := context.WithValue(req.Context(), emailContextKey, u.Email)
	req = req.WithContext(c)

	http.HandlerFunc(routes.ListSchedules).ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Status was not OK: %d\n", status)
	}

	schedules := []Schedule{}
	json.NewDecoder(rr.Body).Decode(&schedules)

	if len(schedules) != 2 {
		t.Errorf("Not the right number of schedules! expected %d got %d", 2, len(schedules))
	}
}

func TestMeRoute(t *testing.T) {
	// create dummy db
	f, _ := ioutil.TempFile("", "")
	db, err := gorm.Open("sqlite3", f.Name())
	defer os.Remove(f.Name())
	defer db.Close()

	if err != nil {
		t.Fatal("Error initializing test sqlite db")
	}

	// initialize routes
	routes := NewRoutes(db, []byte{}, &HTTPClient{})
	routes.MigrateDB()

	password := "s3cure_pw"
	// create dummy user
	u := User{
		Email: "dummyuser",
		Hash:  createPasswordHash(password),
	}

	db.Create(&u)

	t.Run("returns the user", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/me", nil)
		c := context.WithValue(req.Context(), emailContextKey, u.Email)
		req = req.WithContext(c)
		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(routes.Me)
		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Status was not 200 OK: %d\n", status)
		}
	})
}
