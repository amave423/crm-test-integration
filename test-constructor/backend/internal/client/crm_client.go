package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CRMClient interface {
	GetEvents() ([]Event, error)
	GetEventSpecializations(eventID int) (*EventDetailResponse, error)
	CreateTestSession(applicationID, testID uint, sessionID string, expiresAt time.Time) error
	SendTestResult(applicationID uint, result CRMResultData) error
	ExchangeTicket(ticket string) (*CRMSSOExchangeResponse, error)
	SendApplicationStatus(applicationID uint, status string) error
}

type crmClient struct {
	baseURL string
	token   string
	client  *http.Client
}

func NewCRMClient(baseURL, token string) CRMClient {
	return &crmClient{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

type Event struct {
	ID              int              `json:"id"`
	Name            string           `json:"name"`
	Title           string           `json:"title,omitempty"`
	StartDate       string           `json:"start_date"`
	StartDateAlt    string           `json:"startDate,omitempty"`
	EndDate         string           `json:"end_date"`
	EndDateAlt      string           `json:"endDate,omitempty"`
	Specializations []Specialization `json:"specializations"`
}

type Specialization struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
}

type EventDetailResponse struct {
	Specializations []Specialization `json:"specializations"`
}

type CRMCreateSessionData struct {
	TestID    uint   `json:"test_id"`
	SessionID string `json:"session_id"`
	ExpiresAt string `json:"expires_at"`
}

type CRMSSOUser struct {
	ID                uint   `json:"id"`
	Email             string `json:"email"`
	FirstName         string `json:"first_name"`
	LastName          string `json:"last_name"`
	DisplayName       string `json:"display_name"`
	Role              string `json:"role"`
	ManagedEventIDs   []uint `json:"managed_event_ids"`
	IsGlobalOrganizer bool   `json:"is_global_organizer"`
	VK                string `json:"vk"`
	VKConfirmed       bool   `json:"vk_confirmed"`
	Course            *int   `json:"course"`
	Specialty         string `json:"specialty"`
	Specializations   any    `json:"specializations"`
}

type CRMContextApplication struct {
	ID int `json:"id"`
}

type CRMContextEvent struct {
	ID int `json:"id"`
}

type CRMContextSpecialization struct {
	ID int `json:"id"`
}

type CRMTestingContext struct {
	Application    CRMContextApplication     `json:"application"`
	Event          *CRMContextEvent          `json:"event"`
	Specialization *CRMContextSpecialization `json:"specialization"`
	AvailableTests any                       `json:"availableTests"`
	CurrentSession any                       `json:"currentSession"`
	LatestResult   any                       `json:"latestResult"`
}

type CRMSSOExchangeResponse struct {
	User        CRMSSOUser         `json:"user"`
	Application *CRMTestingContext `json:"application"`
	Next        string             `json:"next"`
}

func (c *crmClient) GetEvents() ([]Event, error) {
	url := c.baseURL + "/api/users/events/"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-Service-Token", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к CRM: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CRM вернул ошибку %d: %s", resp.StatusCode, string(body))
	}

	var events []Event
	if err := json.Unmarshal(body, &events); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	return events, nil
}

func (c *crmClient) GetEventSpecializations(eventID int) (*EventDetailResponse, error) {
	url := c.baseURL + fmt.Sprintf("/api/users/events/%d/", eventID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("X-Service-Token", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к CRM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("event not found")
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("CRM вернул ошибку %d: %s", resp.StatusCode, string(body))
	}

	var eventData EventDetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&eventData); err != nil {
		return nil, fmt.Errorf("ошибка декодирования ответа: %w", err)
	}

	return &eventData, nil
}

func (c *crmClient) CreateTestSession(applicationID, testID uint, sessionID string, expiresAt time.Time) error {
	url := c.baseURL + fmt.Sprintf("/api/users/integration/applications/%d/test-sessions/", applicationID)

	data := CRMCreateSessionData{
		TestID:    testID,
		SessionID: sessionID,
		ExpiresAt: expiresAt.Format("2006-01-02T15:04:05Z"),
	}

	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга данных: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("CRM вернул ошибку %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

func (c *crmClient) SendTestResult(applicationID uint, result CRMResultData) error {
	url := c.baseURL + fmt.Sprintf("/api/users/integration/applications/%d/test-results/", applicationID)

	body, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга данных: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("ошибка создания запроса: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка отправки результатов: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errResp map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("CRM вернул ошибку %d: %v", resp.StatusCode, errResp)
	}

	return nil
}

func (c *crmClient) ExchangeTicket(ticket string) (*CRMSSOExchangeResponse, error) {
	body, err := json.Marshal(map[string]string{"ticket": ticket})
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга тикета: %w", err)
	}

	url := c.baseURL + "/api/users/integration/testing/sso-exchange/"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса SSO: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Service-Token", c.token)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса SSO к CRM: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа SSO: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("CRM SSO exchange failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	var payload CRMSSOExchangeResponse
	if err := json.Unmarshal(responseBody, &payload); err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа SSO: %w", err)
	}

	return &payload, nil
}

type CRMResultData struct {
	SessionID         string `json:"session_id"`
	TestID            string `json:"test_id,omitempty"`
	Score             int    `json:"score"`
	MaxScore          int    `json:"max_score"`
	IsPassed          bool   `json:"is_passed"`
	CompletedAt       string `json:"completed_at"`
	StartedAt         string `json:"started_at"`
	ApplicationStatus string `json:"application_status,omitempty"`
}

func (c *crmClient) SendApplicationStatus(applicationID uint, status string) error {
	url := c.baseURL + fmt.Sprintf("/api/users/integration/applications/%d/status/", applicationID)
	payload := map[string]string{"application_status": status}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("status data marshal failed: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadJSON))
	if err != nil {
		return fmt.Errorf("request creation failed: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Service-Token", c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("CRM request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("CRM returned %d: %s", resp.StatusCode, string(responseBody))
	}

	return nil
}
