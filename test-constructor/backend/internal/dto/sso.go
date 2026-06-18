package dto

type SSOExchangeRequest struct {
	Ticket string `json:"ticket"`
}

type SSOExchangeResponse struct {
	Token       string      `json:"token"`
	UserID      uint        `json:"user_id"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	Surname     string      `json:"surname"`
	Role        string      `json:"role"`
	Message     string      `json:"message"`
	Application interface{} `json:"application,omitempty"`
	Next        string      `json:"next,omitempty"`
	TestLink    string      `json:"test_link,omitempty"`
}
