package database

import (
	"github.com/g-linville/budgeting/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// InitDB initializes the database connection and runs auto-migrations
func InitDB(dbPath string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Enable foreign key constraints (SQLite requires explicit enablement)
	db.Exec("PRAGMA foreign_keys = ON;")

	// Auto-migrate schema - ORDER MATTERS (parent tables first)
	err = db.AutoMigrate(
		&models.Category{},
		&models.RecurringExpense{},
		&models.RecurringIncome{},
		&models.Expense{},
		&models.Income{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Ping checks database connectivity
func Ping(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
