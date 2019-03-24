package api

import (
	"crypto/rand"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"

	"github.com/jinzhu/gorm"
)

// Routes is used to pass dependencies in to each of the routes
type Routes struct {
	db         *gorm.DB
	jwtSecret  []byte
	httpClient *HTTPClient
}

// NewRoutes constructs a new Routes object with the require deps. If jwtSecret is empty
// a random one will be generated.
func NewRoutes(db *gorm.DB, jwtSecret []byte, httpClient *HTTPClient) *Routes {
	var secret []byte
	if len(jwtSecret) == 0 {
		secret = make([]byte, 32)
		rand.Read(secret)
	} else {
		secret = jwtSecret
	}

	return &Routes{
		db:         db,
		jwtSecret:  secret,
		httpClient: httpClient,
	}
}

// Me returns the currently authenticated user
func (routes *Routes) Me(w http.ResponseWriter, r *http.Request) {
	email := r.Context().Value(emailContextKey).(string)

	u := User{}
	routes.db.Where("email = ?", email).First(&u)

	b, _ := json.Marshal(u)

	w.Write(b)
}

// CreateSchedule creates a schedule. Uses the user's name as the Source
func (routes *Routes) CreateSchedule(w http.ResponseWriter, r *http.Request) {
	email := r.Context().Value(emailContextKey).(string)

	timeString := r.FormValue("time")
	if strings.TrimSpace(timeString) == "" {
		writeErrorMessage(w, "Time is required", http.StatusBadRequest)
		return
	}

	time, err := time.Parse(time.RFC3339, timeString)
	if err != nil {
		writeErrorMessage(w, "Invalid time format. Must be format RFC3339", http.StatusBadRequest)
		return
	}

	user := User{}
	routes.db.First(&user, "email = ?", email)

	sched := Schedule{
		Time:   time,
		Source: user.Name,
		Status: "PENDING",
	}

	routes.db.Create(&sched)
	w.WriteHeader(http.StatusCreated)
}

// ListSchedules returns a list of schedules
func (routes *Routes) ListSchedules(w http.ResponseWriter, r *http.Request) {
	var schedules []Schedule
	routes.db.Find(&schedules)

	b, err := json.Marshal(schedules)
	if err != nil {
		writeErrorMessage(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write(b)
}

// DeleteSchedule deletes the schedule from the db
func (routes *Routes) DeleteSchedule(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]

	s := Schedule{}
	routes.db.Where("id = ?", id).First(&s)

	if s.ID == 0 {
		writeErrorMessage(w, "Not Found", http.StatusNotFound)
		return
	}

	routes.db.Delete(&s)
}

type message struct {
	Message string `json:"message"`
}

func writeErrorMessage(w http.ResponseWriter, m string, code int) {
	p := message{Message: m}
	s, _ := json.Marshal(p)

	w.Header().Add("Content-Type", "application/json")
	http.Error(w, string(s), code)
}
