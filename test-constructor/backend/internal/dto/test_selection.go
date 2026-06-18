package dto

type TestSelectionResponse struct {
	EventID          uint                `json:"event_id"`
	SpecializationID uint                `json:"specialization_id"`
	ApplicationID    uint                `json:"application_id"`
	Tests            []TestSelectionInfo `json:"tests"`
	AllCompleted     bool                `json:"all_completed"`
	EventPassed      bool                `json:"event_passed"`
}

type TestSelectionInfo struct {
	ConfigID    uint   `json:"config_id"`
	TestID      uint   `json:"test_id"`
	TestLink    string `json:"test_link"`
	Title       string `json:"title"`
	Description string `json:"description"`
	TimeLimit   int    `json:"time_limit"`
	Status      string `json:"status"`
	Score       int    `json:"score,omitempty"`
	MaxScore    int    `json:"max_score,omitempty"`
	Passed      bool   `json:"passed,omitempty"`
	AttemptID   uint   `json:"attempt_id,omitempty"`
	IsExtra     bool   `json:"is_extra"`
	Message     string `json:"message,omitempty"`
}
