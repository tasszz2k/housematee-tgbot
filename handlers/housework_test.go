package handlers

import (
	"strings"
	"testing"
	"time"

	"housematee-tgbot/models"
)

func TestDueStatusText(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		t.Fatalf("failed to load location: %s", err.Error())
	}
	now := time.Now().In(loc)

	testCases := []struct {
		name     string
		date     string
		expected string
	}{
		{name: "today", date: now.Format("02/01/2006"), expected: "today"},
		{name: "tomorrow", date: now.AddDate(0, 0, 1).Format("02/01/2006"), expected: "in 1 day"},
		{name: "in 3 days", date: now.AddDate(0, 0, 3).Format("02/01/2006"), expected: "in 3 days"},
		{name: "yesterday", date: now.AddDate(0, 0, -1).Format("02/01/2006"), expected: "overdue 1 day"},
		{name: "overdue 4 days", date: now.AddDate(0, 0, -4).Format("02/01/2006"), expected: "overdue 4 days"},
		{name: "invalid", date: "", expected: ""},
	}

	for _, tc := range testCases {
		if actual := DueStatusText(tc.date); actual != tc.expected {
			t.Errorf("%s: expected %q, got %q", tc.name, tc.expected, actual)
		}
	}
}

func TestConvertHouseworkToMarkdownFormatEmptyFields(t *testing.T) {
	text := ConvertHouseworkToMarkdownFormat(models.Task{Name: "Trash", Frequency: 1, Assignee: "@a"})
	for _, want := range []string{"*Trash*", "Assignee: @a", "Every 1 day", "Last done: -", "Next due: -", "Note: -"} {
		if !strings.Contains(text, want) {
			t.Errorf("text %q does not contain %q", text, want)
		}
	}
}
