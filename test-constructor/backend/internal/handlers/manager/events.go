package manager

import (
	"encoding/json"
	"io"
	"net/http"
	"test-constructor/config"
)

type Event struct {
	ID              uint             `json:"id"`
	Name            string           `json:"name"`
	Title           string           `json:"title,omitempty"`
	StartDate       string           `json:"start_date"`
	StartDateAlt    string           `json:"startDate,omitempty"`
	EndDate         string           `json:"end_date"`
	EndDateAlt      string           `json:"endDate,omitempty"`
	Specializations []Specialization `json:"specializations"`
}

// @Summary Получить мероприятия
// @Security ApiKeyAuth
// @Description Получить список мероприятий
// @Tags manager
// @Produce json
// @Success 200 {object} Event
// @Failure 404 {object} map[string]string
// @Router /api/manager/events [get]
func GetEvents(w http.ResponseWriter, r *http.Request) {
	cfg := config.Load()
	crmService := cfg.CRMService
	crmToken := cfg.CRMToken
	req, err := http.NewRequest("GET", crmService+"/api/users/events/", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req.Header.Set("X-Service-Token", crmToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if resp.StatusCode != http.StatusOK {
		http.Error(w, string(body), resp.StatusCode)
		return
	}

	var events []Event
	err = json.Unmarshal(body, &events)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for index := range events {
		if events[index].Name == "" {
			events[index].Name = events[index].Title
		}
		if events[index].StartDate == "" {
			events[index].StartDate = events[index].StartDateAlt
		}
		if events[index].EndDate == "" {
			events[index].EndDate = events[index].EndDateAlt
		}
		for specIndex := range events[index].Specializations {
			if events[index].Specializations[specIndex].Name == "" {
				events[index].Specializations[specIndex].Name = events[index].Specializations[specIndex].Title
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(events)
}
