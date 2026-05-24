package models

import (
    "gorm.io/gorm"
)

type Test struct {
    gorm.Model
    CreatorID    uint    `json:"creator_id"`
    Title        string  `json:"title"`
    Description  string  `json:"description"`
    IsExtra      bool    `json:"is_extra"`
    IsPercentage bool    `json:"is_percentage"`
    Threshold    float64 `json:"threshold"`
    SuccessText  string  `json:"success_text"`
    FailText     string  `json:"fail_text"`
    CompleteTime int     `json:"complete_time"`

    Questions []Question
    User      User `gorm:"foreignKey:CreatorID;constraint:OnDelete:CASCADE;OnUpdate:CASCADE"`
}
