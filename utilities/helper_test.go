package utilities

import (
	"testing"
	"time"
)

func TestDaysUntil(t *testing.T) {
	testCases := []struct {
		name   string
		offset int
	}{
		{name: "today", offset: 0},
		{name: "tomorrow", offset: 1},
		{name: "next week", offset: 7},
		{name: "yesterday", offset: -1},
		{name: "last month", offset: -30},
	}

	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		t.Fatalf("failed to load location: %s", err.Error())
	}

	for _, tc := range testCases {
		date := time.Now().In(loc).AddDate(0, 0, tc.offset).Format("02/01/2006")
		actual, err := DaysUntil(date)
		if err != nil {
			t.Fatalf("%s: unexpected error: %s", tc.name, err.Error())
		}
		if actual != tc.offset {
			t.Errorf("%s: expected %d, got %d", tc.name, tc.offset, actual)
		}
	}

	if _, err := DaysUntil("not a date"); err == nil {
		t.Error("expected an error for an invalid date")
	}
}

func TestIsDateDueOrOverdue(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{

		{
			name:     "overdue",
			input:    "01/09/2000",
			expected: true, // Overdue, before the current date
		},
		{
			name:     "due",
			input:    "11/09/2023",
			expected: true, // Due on or before the current date
		},
		{
			name:     "false",
			input:    "01/01/2099",
			expected: false, // Future date
		},
	}

	for _, testCase := range testCases {
		// Act
		actual, err := IsDateDueOrOverdue(testCase.input)
		if err != nil {
			t.Errorf("unexpected error: %s", err.Error())
		}
		if actual != testCase.expected {
			t.Errorf("%s: expected %v, got %v", testCase.name, testCase.expected, actual)
		}
	}
}
