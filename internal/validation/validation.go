package validation

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/g-linville/budgeting/internal/utils"
)

// ValidationError represents a single validation error
type ValidationError struct {
	Field   string
	Message string
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return "no validation errors"
	}

	var messages []string
	for _, err := range v {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}
	return strings.Join(messages, "; ")
}

// HasErrors returns true if there are any validation errors
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// ValidateExpense validates expense input data
// Returns the amount in cents and any validation errors
func ValidateExpense(name, amountStr, dateStr string) (int, time.Time, ValidationErrors) {
	var errors ValidationErrors

	// Validate name
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Name is required",
		})
	} else if len(trimmedName) > 255 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Name must be 255 characters or less",
		})
	}

	// Validate and convert amount
	var amountCents int
	if strings.TrimSpace(amountStr) == "" {
		errors = append(errors, ValidationError{
			Field:   "amount",
			Message: "Amount is required",
		})
	} else {
		cents, err := utils.DollarsToCents(amountStr)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   "amount",
				Message: "Amount must be a positive number",
			})
		} else {
			amountCents = cents
		}
	}

	// Validate date
	var date time.Time
	if strings.TrimSpace(dateStr) == "" {
		// Default to today if not provided
		date = time.Now()
	} else {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   "expense_date",
				Message: "Invalid date format (use YYYY-MM-DD)",
			})
		} else {
			date = parsedDate
		}
	}

	return amountCents, date, errors
}

// ValidateIncome validates income input data
// Returns the amount in cents and any validation errors
func ValidateIncome(name, amountStr, dateStr string) (int, time.Time, ValidationErrors) {
	// Income validation is identical to expense validation (just no category)
	return ValidateExpense(name, amountStr, dateStr)
}

// ValidateCategory validates category input data
func ValidateCategory(name, color string) ValidationErrors {
	var errors ValidationErrors

	// Validate name
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Category name is required",
		})
	} else if len(trimmedName) > 255 {
		errors = append(errors, ValidationError{
			Field:   "name",
			Message: "Category name must be 255 characters or less",
		})
	}

	// Validate color (optional)
	trimmedColor := strings.TrimSpace(color)
	if trimmedColor != "" {
		// Validate hex color format (#RRGGBB)
		matched, _ := regexp.MatchString("^#[0-9A-Fa-f]{6}$", trimmedColor)
		if !matched {
			errors = append(errors, ValidationError{
				Field:   "color",
				Message: "Color must be in hex format (e.g., #FF5733)",
			})
		}
	}

	return errors
}

// ValidateName validates a name field (generic)
func ValidateName(name string) error {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return fmt.Errorf("name is required")
	}
	if len(trimmed) > 255 {
		return fmt.Errorf("name too long (max 255 characters)")
	}
	return nil
}
