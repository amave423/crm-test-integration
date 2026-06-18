package dto

type CreateUserEventRequest struct {
	EventID          uint `json:"event_id"`
	SpecializationID uint `json:"specialization_id"`
	ApplicationID    uint `json:"application_id"`
}

type UserEventResponse struct {
	ID               uint `json:"id"`
	EventID          uint `json:"event_id"`
	SpecializationID uint `json:"specialization_id"`
	ApplicationID    uint `json:"application_id"`
}

type UserEventsListResponse struct {
	Events []UserEventResponse `json:"events"`
}
