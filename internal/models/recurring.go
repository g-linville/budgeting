package models

import "time"

type RecurringExpense struct {
	ID         uint       `gorm:"primaryKey"`
	Name       string     `gorm:"not null"`
	Amount     int        `gorm:"not null"` // Stored as cents
	CategoryID *uint
	Cadence    string     `gorm:"not null"` // 'monthly', 'semi-annual', 'annual'
	StartDate  time.Time  `gorm:"type:date;not null"`
	NextDate   time.Time  `gorm:"type:date;index;not null"`
	EndDate    *time.Time `gorm:"type:date"` // Nullable
	Active     bool       `gorm:"default:true"`
	CreatedAt  time.Time  `gorm:"autoCreateTime"`

	// Relationships
	Category *Category `gorm:"foreignKey:CategoryID;constraint:OnDelete:SET NULL"`
	Expenses []Expense `gorm:"foreignKey:RecurringID"`
}

type RecurringIncome struct {
	ID        uint       `gorm:"primaryKey"`
	Name      string     `gorm:"not null"`
	Amount    int        `gorm:"not null"` // Stored as cents
	Cadence   string     `gorm:"not null"` // 'monthly', 'semi-annual', 'annual'
	StartDate time.Time  `gorm:"type:date;not null"`
	NextDate  time.Time  `gorm:"type:date;index;not null"`
	EndDate   *time.Time `gorm:"type:date"` // Nullable
	Active    bool       `gorm:"default:true"`
	CreatedAt time.Time  `gorm:"autoCreateTime"`

	// Relationships
	Incomes []Income `gorm:"foreignKey:RecurringID"`
}
