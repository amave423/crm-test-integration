package dto

type StatisticsFilter struct {
	IsExtra *bool `json:"is_extra,omitempty"`
}

type StatisticsResponse struct {
	Attempts []UserAttemptInfo `json:"attempts"`
}

type UserAttemptInfo struct {
	UserID    uint               `json:"user_id"`
	FirstName string             `json:"first_name"`
	LastName  string             `json:"last_name"`
	Email     string             `json:"email"`
	Score     int                `json:"score"`
	MaxScore  int                `json:"max_score"`
	Passed    bool               `json:"passed"`
	TimeSpent int                `json:"time_spent_minutes"`
	IsExtra   bool               `json:"is_extra"`
	Questions []QuestionStatInfo `json:"questions"`
}

type QuestionStatInfo struct {
	Text         string `json:"text"`
	Points       int    `json:"points_earned"`
	MaxPoints    int    `json:"max_points"`
	IsCorrect    bool   `json:"is_correct"`
	QuestionType string `json:"question_type"`
	OrderNumber  int    `json:"order_number"`
}

type UserStatisticsResponse struct {
	UserID    uint                `json:"user_id"`
	FirstName string              `json:"first_name"`
	LastName  string              `json:"last_name"`
	Email     string              `json:"email"`
	Attempts  []UserAttemptDetail `json:"attempts"`
}

type UserAttemptDetail struct {
	AttemptID uint               `json:"attempt_id"`
	TestTitle string             `json:"test_title"`
	EventName string             `json:"event_name"`
	IsExtra   bool               `json:"is_extra"`
	Score     int                `json:"score"`
	MaxScore  int                `json:"max_score"`
	Passed    bool               `json:"passed"`
	Questions []QuestionStatInfo `json:"questions"`
}

type GetUsersResponse struct {
	Users []UserInfo `json:"users"`
}

type UserInfo struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Surname string `json:"surname"`
	Email   string `json:"email"`
}
