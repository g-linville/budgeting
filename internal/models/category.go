package models

import "time"

type Category struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"uniqueIndex;not null"`
	Color     string    `gorm:"size:7"`
	CreatedAt time.Time `gorm:"autoCreateTime"`

	// Relationships
	Expenses          []Expense          `gorm:"foreignKey:CategoryID"`
	RecurringExpenses []RecurringExpense `gorm:"foreignKey:CategoryID"`
}
