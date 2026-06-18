package dto

import "test-constructor/internal/domain"

type CreateTestRequest struct {
	Title       string                  `json:"title"`
	Description string                  `json:"description"`
	Questions   []CreateQuestionRequest `json:"questions"`
}

type CreateQuestionRequest struct {
	Text        string                 `json:"text"`
	Points      int                    `json:"points"`
	Type        string                 `json:"type"`
	OrderNumber int                    `json:"order_number"`
	Options     domain.QuestionOptions `json:"options"`
}

type TestResponse struct {
	ID          uint   `json:"test_id"`
	CreatorID   uint   `json:"creator_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatorName string `json:"creator_name,omitempty"`
}

type TestsListResponse struct {
	Tests []TestResponse `json:"tests"`
}

type CreateTestResponse struct {
	ID      uint   `json:"id"`
	Message string `json:"message"`
}

type TestDetailResponse struct {
	TestResponse
	Questions []QuestionResponse `json:"questions"`
}

type QuestionResponse struct {
	ID          uint                   `json:"id"`
	Text        string                 `json:"text"`
	Points      int                    `json:"points"`
	Type        string                 `json:"type"`
	OrderNumber int                    `json:"order_number"`
	Options     domain.QuestionOptions `json:"options"`
}

type DeleteTestResponse struct {
	Message string `json:"message" example:"Тест удален"`
}
