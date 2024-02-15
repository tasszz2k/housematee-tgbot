package utilities

import (
	"strconv"
	"strings"
)

// ParseAmount processes the amount string and converts "k" and "m" to their respective multipliers.
func ParseAmount(amountStr string) string {
	if strings.Contains(amountStr, "k") {
		multiplier := 1000
		value, err := strconv.ParseFloat(strings.Replace(amountStr, "k", "", 1), 64)
		if err != nil {
			return ""
		}
		return strconv.FormatFloat(value*float64(multiplier), 'f', 0, 64)
	} else if strings.Contains(amountStr, "m") {
		multiplier := 1000000
		value, err := strconv.ParseFloat(strings.Replace(amountStr, "m", "", 1), 64)
		if err != nil {
			return ""
		}
		return strconv.FormatFloat(value*float64(multiplier), 'f', 0, 64)
	}
	return amountStr
}

// FormatMoney Format number to money format
// e.g., 100000 -> 100,000 ₫
func FormatMoney(amount int) string {
	amountStr := strconv.FormatUint(uint64(amount), 10)
	var result string

	// Iterate over the amount string from the end and add commas after every third digit
	for i := len(amountStr) - 1; i >= 0; i-- {
		if (len(amountStr)-i-1)%3 == 0 && i != len(amountStr)-1 {
			result = "," + result
		}
		result = string(amountStr[i]) + result
	}

	return result + " ₫"
}
