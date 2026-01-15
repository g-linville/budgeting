package utils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// DollarsToCents converts a dollar string to cents (integer)
// Examples: "12.34" -> 1234, "5" -> 500, "0.99" -> 99, "$1,000.50" -> 100050
func DollarsToCents(dollars string) (int, error) {
	// Clean input: trim whitespace, remove "$" and ","
	cleaned := strings.TrimSpace(dollars)
	cleaned = strings.ReplaceAll(cleaned, "$", "")
	cleaned = strings.ReplaceAll(cleaned, ",", "")

	// Handle empty string
	if cleaned == "" {
		return 0, fmt.Errorf("amount cannot be empty")
	}

	// Parse as float
	value, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount format: %w", err)
	}

	// Validate positive
	if value <= 0 {
		return 0, fmt.Errorf("amount must be positive")
	}

	// Convert to cents (multiply by 100, round to nearest integer)
	cents := int(math.Round(value * 100))

	return cents, nil
}

// CentsToUSD converts cents (integer) to formatted USD string
// Examples: 1234 -> "$12.34", 500000 -> "$5,000.00", 99 -> "$0.99", -1234 -> "-$12.34"
func CentsToUSD(cents int) string {
	negative := cents < 0
	if negative {
		cents = -cents
	}

	dollars := cents / 100
	remainder := cents % 100

	// Format dollars with thousand separators
	dollarStr := strconv.Itoa(dollars)
	var result strings.Builder

	for i, c := range dollarStr {
		if i > 0 && (len(dollarStr)-i)%3 == 0 {
			result.WriteByte(',')
		}
		result.WriteRune(c)
	}

	sign := ""
	if negative {
		sign = "-"
	}
	return fmt.Sprintf("%s$%s.%02d", sign, result.String(), remainder)
}
