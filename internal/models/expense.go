package models

import "time"

type Expense struct {
	ID          uint      `gorm:"primaryKey"`
	Name        string    `gorm:"not null"`
	Amount      int       `gorm:"not null"` // Stored as cents
	CategoryID  *uint     `gorm:"index"`    // Nullable FK
	ExpenseDate time.Time `gorm:"type:date;index;not null"`
	Notes       string    `gorm:"type:text"`
	RecurringID *uint     `gorm:"index"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`

	// Relationships
	Category         *Category         `gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL"`
	RecurringExpense *RecurringExpense `gorm:"foreignKey:RecurringID;constraint:OnDelete:SET NULL"`
}
