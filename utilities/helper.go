package utilities

import (
	"google.golang.org/genproto/googleapis/type/date"
	"time"
)

func GetCurrentDate() string {
	return time.Now().Format("02/01/2006")
}

// AddDay add day operation
func AddDay(dateStr string, day int) (string, error) {
	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return "", err
	}
	t = t.AddDate(0, 0, day)
	return t.Format("02/01/2006"), nil
}

func StringToGoogleDate(dateStr string) (date.Date, error) {
	t, err := time.Parse("02/01/2006", dateStr)
	if err != nil {
		return date.Date{}, err
	}
	return DateToGoogleDate(t), nil
}

func DateToGoogleDate(t time.Time) date.Date {
	return date.Date{
		Year:  int32(t.Year()),
		Month: int32(t.Month()),
		Day:   int32(t.Day()),
	}
}

func IsDateDueOrOverdue(dateStr string) (bool, error) {
	// Load the Asia/Bangkok time zone (GMT+7)
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return false, err
	}

	// Parse the date string with the specified time zone
	t, err := time.ParseInLocation("02/01/2006", dateStr, loc)
	if err != nil {
		return false, err
	}

	// Get the current date in the same time zone
	currentDate := time.Now().In(loc)

	// Compare the parsed date with the current date
	if t.Before(currentDate) || t.Equal(currentDate) {
		// The date is due or overdue
		return true, nil
	}

	// The date is in the future
	return false, nil
}
