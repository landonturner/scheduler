package api

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
)

func TestCheckSchedules(t *testing.T) {
	// create dummy db
	f, _ := ioutil.TempFile("", "")
	db, err := gorm.Open("sqlite3", f.Name())
	defer os.Remove(f.Name())
	defer db.Close()

	if err != nil {
		t.Fatal("Error initializing test sqlite db")
	}

	t.Run("client success", func(t *testing.T) {
		// create web endpoint
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte(`OK`))
		}))
		// Close the server when test finishes
		defer server.Close()
		httpClient := HTTPClient{
			client: server.Client(),
			url:    server.URL,
		}

		// initialize routes
		routes := NewRoutes(db, []byte{}, &httpClient)
		routes.MigrateDB()

		oneMinuteAgo := time.Now().Add(-1 * time.Minute)
		oneMinuteInFuture := time.Now().Add(time.Minute)

		routes.db.Create(&Schedule{
			Time:   oneMinuteAgo,
			Status: "PENDING",
			Source: "jimbobjoe",
		})
		routes.db.Create(&Schedule{
			Time:   oneMinuteInFuture,
			Status: "PENDING",
			Source: "jimbobjoe",
		})

		routes.CheckSchedules()

		sentSchedule := Schedule{}
		routes.db.Where("status = ?", "SENT").First(&sentSchedule)
		if sentSchedule.ID == 0 {
			t.Error("Did not set schedule status to sent")
		}

		pendingSchedule := Schedule{}
		routes.db.Where("status = ?", "PENDING").First(&pendingSchedule)
		if pendingSchedule.ID == 0 {
			t.Error("Incorrectly affected pending schedule")
		}
	})

	t.Run("client errors", func(t *testing.T) {
		// create fake web server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(401)
		}))

		// Close the server when test finishes
		defer server.Close()
		httpClient := HTTPClient{
			client: server.Client(),
			url:    server.URL,
		}

		// initialize routes
		routes := NewRoutes(db, []byte{}, &httpClient)
		routes.MigrateDB()

		oneMinuteAgo := time.Now().Add(-1 * time.Minute)

		routes.db.Create(&Schedule{
			Time:   oneMinuteAgo,
			Status: "PENDING",
			Source: "jimbobjoe",
		})

		routes.CheckSchedules()

		pendingSchedule := Schedule{}
		routes.db.Where("status = ?", "ERROR").First(&pendingSchedule)
		if pendingSchedule.ID == 0 {
			t.Error("Did not mark the schedule as error")
		}
	})
}

func TestExecuteSchedule(t *testing.T) {
	executed := false
	// create fake web endpoint
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		executed = true
		w.Write([]byte("ok"))
	}))
	defer server.Close()

	httpClient := HTTPClient{
		client: server.Client(),
		url:    server.URL,
	}

	httpClient.executeSchedule()

	if !executed {
		t.Error("HTTP Endpoint not executed")
	}
}
