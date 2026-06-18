package domain

type ExtraThreshold struct {
	ExtraThresholdID uint `gorm:"primaryKey"`
	ConfigID         uint `gorm:"not null"`
	TestID           uint `gorm:"not null"`
	Threshold        int  `gorm:"not null"`
	Message          string
	ExtraConfig      EventConfig `gorm:"foreignKey:ExtraConfigID;constraint:OnDelete:CASCADE;"`
	ExtraConfigID    uint
}

func (ExtraThreshold) TableName() string {
	return "extra_thresholds"
}
