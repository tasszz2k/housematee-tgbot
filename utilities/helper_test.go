package utilities

import (
	"testing"
)

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
