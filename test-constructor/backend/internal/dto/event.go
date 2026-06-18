package dto

type EventResponse struct {
	ID              uint                          `json:"id"`
	Name            string                        `json:"name"`
	StartDate       string                        `json:"start_date"`
	EndDate         string                        `json:"end_date"`
	Specializations []EventSpecializationResponse `json:"specializations"`
}

type EventsListResponse struct {
	Events []EventResponse `json:"events"`
	Total  int             `json:"total"`
}

type EventSpecializationResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EventSpecializationsListResponse struct {
	EventID         int                           `json:"event_id"`
	Specializations []EventSpecializationResponse `json:"specializations"`
}
