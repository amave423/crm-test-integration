package dto

import "test-constructor/internal/domain"

type StartAttemptRequest struct {
	ApplicationID uint `json:"application_id"`
}

type StartAttemptResponse struct {
	ConfigID      uint             `json:"config_id"`
	TestID        uint             `json:"test_id"`
	ApplicationID uint             `json:"application_id"`
	Title         string           `json:"title"`
	Description   string           `json:"description"`
	TimeLimit     int              `json:"time_limit"`
	Threshold     int              `json:"threshold"`
	Questions     []QuestionPublic `json:"questions"`
}

type QuestionPublic struct {
	QuestionID  uint          `json:"question_id"`
	Text        string        `json:"text"`
	Points      int           `json:"points"`
	OrderNumber int           `json:"order_number"`
	Type        domain.QType  `json:"type"`
	Options     PublicOptions `json:"options"`
}

type PublicOptions struct {
	Choices       []PublicChoice  `json:"choice,omitempty"`
	Matching      *PublicMatching `json:"matching,omitempty"`
	CaseSensitive bool            `json:"case_sensitive,omitempty"`
	Sequence      []string        `json:"sequence,omitempty"`
}

type PublicChoice struct {
	Text  string `json:"text"`
	Index int    `json:"index"`
}

type PublicMatching struct {
	LeftColumn  []string `json:"left,omitempty"`
	RightColumn []string `json:"right,omitempty"`
}

type FinishAttemptRequest struct {
	UserAnswers []UserAnswerInfo `json:"user_answers"`
}

type UserAnswerInfo struct {
	QuestionID uint       `json:"question_id"`
	Answer     UserAnswer `json:"answer"`
}

type UserAnswer struct {
	Choices       []bool         `json:"choices,omitempty"`
	MatchingPairs []MatchingPair `json:"matching,omitempty"`
	UserInput     string         `json:"user_input,omitempty"`
	Sequence      []SequenceItem `json:"sequence,omitempty"`
}

type MatchingPair struct {
	Left  string `json:"left"`
	Right string `json:"right"`
}

type SequenceItem struct {
	Text  string `json:"text"`
	Order int    `json:"order"`
}

type FinishAttemptResponse struct {
	Result        string `json:"result"`
	Score         int    `json:"score"`
	MaxTestPoints int    `json:"max_test_points"`
	Passed        bool   `json:"passed"`
	AllCompleted  bool   `json:"all_completed"`
}
