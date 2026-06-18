package dto

type CreateEventConfigRequest struct {
	EventID          uint                    `json:"event_id"`
	SpecializationID uint                    `json:"specialization_id"`
	TestID           uint                    `json:"test_id"`
	SuccessText      string                  `json:"success_text"`
	FailText         string                  `json:"fail_text"`
	TimeLimit        int                     `json:"time_limit"`
	Threshold        int                     `json:"threshold"`
	ExtraThreshold   []ExtraThresholdRequest `json:"extra_threshold"`
}

type UpdateEventConfigRequest struct {
	EventID          uint                    `json:"event_id"`
	SpecializationID uint                    `json:"specialization_id"`
	TestID           uint                    `json:"test_id"`
	SuccessText      string                  `json:"success_text"`
	FailText         string                  `json:"fail_text"`
	TimeLimit        int                     `json:"time_limit"`
	Threshold        int                     `json:"threshold"`
	ExtraThreshold   []ExtraThresholdRequest `json:"extra_threshold"`
}

type ExtraThresholdRequest struct {
	Threshold int    `json:"threshold"`
	Message   string `json:"message"`
	TestID    uint   `json:"test_id"`
}

type EventConfigResponse struct {
	ConfigID         uint                     `json:"config_id"`
	EventID          uint                     `json:"event_id"`
	SpecializationID uint                     `json:"specialization_id"`
	TestID           uint                     `json:"test_id"`
	SuccessText      string                   `json:"success_text"`
	FailText         string                   `json:"fail_text"`
	TimeLimit        int                      `json:"time_limit"`
	Threshold        int                      `json:"threshold"`
	TestLink         string                   `json:"test_link"`
	ExtraThreshold   []ExtraThresholdResponse `json:"extra_threshold,omitempty"`
}

type ExtraThresholdResponse struct {
	Threshold int    `json:"threshold"`
	Message   string `json:"message"`
	TestID    uint   `json:"test_id"`
}

type EventConfigsResponse struct {
	Configs []EventConfigResponse `json:"configs"`
}

type CreateEventConfigResponse struct {
	ConfigID uint   `json:"config_id"`
	TestLink string `json:"test_link"`
	Created  bool   `json:"created"`
}

type UpdateEventConfigResponse struct {
	ConfigID uint   `json:"config_id"`
	TestLink string `json:"test_link"`
}
