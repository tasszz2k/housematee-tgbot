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
	if len(amountStr) <= 3 {
		return amountStr + " ₫"
	}
	amountStr = amountStr[:len(amountStr)-3] + "," + amountStr[len(amountStr)-3:] + " ₫"
	return amountStr
}
