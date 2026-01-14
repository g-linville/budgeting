package models

import "time"

type Income struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"not null"`
	Amount      int       `gorm:"not null"` // Stored as cents
	IncomeDate  time.Time `gorm:"type:date;index;not null"`
	Notes       string    `gorm:"type:text"`
	RecurringID *uint     `gorm:"index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`

	// Relationships
	RecurringIncome *RecurringIncome `gorm:"foreignKey:RecurringID;constraint:OnDelete:SET NULL"`
}
