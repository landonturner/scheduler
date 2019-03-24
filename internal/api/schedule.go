package api

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

// HTTPClient interacts with the remote endpoint
type HTTPClient struct {
	client *http.Client
	url    string
}

// NewHTTPClient initializes the http client and sets
// the base url to the given url
func NewHTTPClient(url string) *HTTPClient {
	return &HTTPClient{
		url:    url,
		client: http.DefaultClient,
	}
}

// CheckSchedules checks the pending schedules if it is time to deploy and
// calls ExecuteSchedule to deploy if the time has come
func (r *Routes) CheckSchedules() {
	schedules := []Schedule{}
	err := r.db.Where("status = ?", "PENDING").Find(&schedules).Error
	if err != nil {
		log.Printf("Error finding schedules: %s\n", err.Error())
		return
	}

	schedulesToRun := []Schedule{}
	for _, s := range schedules {
		if s.Time.Before(time.Now()) {
			schedulesToRun = append(schedulesToRun, s)
		}
	}

	if len(schedulesToRun) > 0 {
		err := r.httpClient.executeSchedule()
		if err != nil {
			log.Printf("ERROR: %s", err.Error())
			for _, s := range schedulesToRun {
				s.Status = "ERROR"
				if err := r.db.Save(&s).Error; err != nil {
					log.Printf("Error saving status: %s\n", err.Error())
				}
			}
			return
		}

		for _, s := range schedulesToRun {
			s.Status = "SENT"
			if err := r.db.Save(&s).Error; err != nil {
				log.Printf("Error saving status: %s\n", err.Error())
			}
		}
	}
}

// ExecuteSchedule calls the remote endpoint
func (h *HTTPClient) executeSchedule() error {
	log.Println("Executing schedule!")

	resp, err := h.client.Get(h.url)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("Invalid response code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	log.Printf("Response: %s\n", string(body))

	return nil
}
