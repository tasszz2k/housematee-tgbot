package commands

import (
	"strings"
	"testing"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"

	"housematee-tgbot/enum"
	"housematee-tgbot/models"
)

func TestIsAssignee(t *testing.T) {
	testCases := []struct {
		name     string
		user     *gotgbot.User
		assignee string
		expected bool
	}{
		{name: "same username with @", user: &gotgbot.User{Username: "tasszz2k"}, assignee: "@tasszz2k", expected: true},
		{name: "same username without @", user: &gotgbot.User{Username: "tasszz2k"}, assignee: "tasszz2k", expected: true},
		{name: "different case", user: &gotgbot.User{Username: "TassZZ2k"}, assignee: "@tasszz2k", expected: true},
		{name: "spaces around assignee", user: &gotgbot.User{Username: "tasszz2k"}, assignee: " @tasszz2k ", expected: true},
		{name: "other user", user: &gotgbot.User{Username: "ng0cth1nh"}, assignee: "@tasszz2k", expected: false},
		{name: "user without username", user: &gotgbot.User{FirstName: "Tass"}, assignee: "@tasszz2k", expected: false},
		{name: "empty assignee", user: &gotgbot.User{Username: "tasszz2k"}, assignee: "", expected: false},
		{name: "only @", user: &gotgbot.User{Username: "tasszz2k"}, assignee: "@", expected: false},
		{name: "nil user", user: nil, assignee: "@tasszz2k", expected: false},
	}

	for _, tc := range testCases {
		if actual := isAssignee(tc.user, tc.assignee); actual != tc.expected {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, actual)
		}
	}
}

func TestNextMemberAfter(t *testing.T) {
	members := []models.Member{
		{ID: 1, Username: "@a"},
		{ID: 2, Username: "@b"},
		{ID: 3, Username: "@c"},
	}

	testCases := []struct {
		name     string
		members  []models.Member
		current  string
		expected string
	}{
		{name: "middle", members: members, current: "@a", expected: "@b"},
		{name: "wraps around", members: members, current: "@c", expected: "@a"},
		{name: "case insensitive", members: members, current: "@B", expected: "@c"},
		{name: "unknown falls back to first", members: members, current: "@x", expected: "@a"},
		{name: "empty falls back to first", members: members, current: "", expected: "@a"},
		{name: "single member", members: members[:1], current: "@a", expected: "@a"},
		{name: "no members keeps current", members: nil, current: "@a", expected: "@a"},
	}

	for _, tc := range testCases {
		if actual := nextMemberAfter(tc.members, tc.current); actual != tc.expected {
			t.Errorf("%s: expected %q, got %q", tc.name, tc.expected, actual)
		}
	}
}

func TestHouseworkListButtonLabel(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		t.Fatalf("failed to load location: %s", err.Error())
	}
	now := time.Now().In(loc)
	today := now.Format("02/01/2006")
	future := now.AddDate(0, 0, 3).Format("02/01/2006")
	past := now.AddDate(0, 0, -2).Format("02/01/2006")

	testCases := []struct {
		name     string
		task     models.Task
		expected string
	}{
		{
			name:     "future",
			task:     models.Task{Name: "Laundry", Assignee: "@a", NextDue: future},
			expected: enum.EmojiBroom + " Laundry » @a » " + future[:5],
		},
		{
			name:     "today",
			task:     models.Task{Name: "Laundry", Assignee: "@a", NextDue: today},
			expected: enum.EmojiBell + " Laundry » @a » today",
		},
		{
			name:     "overdue",
			task:     models.Task{Name: "Trash", Assignee: "@b", NextDue: past},
			expected: enum.EmojiBell + " Trash » @b » overdue 2d",
		},
		{
			name:     "no assignee and no date",
			task:     models.Task{Name: "Trash"},
			expected: enum.EmojiBroom + " Trash",
		},
		{
			name:     "unparseable date is shown as is",
			task:     models.Task{Name: "Trash", Assignee: "@b", NextDue: "soon"},
			expected: enum.EmojiBroom + " Trash » @b » soon",
		},
	}

	for _, tc := range testCases {
		if actual := houseworkListButtonLabel(tc.task); actual != tc.expected {
			t.Errorf("%s: expected %q, got %q", tc.name, tc.expected, actual)
		}
	}
}

func TestFormatHouseworkAudit(t *testing.T) {
	rotated := formatHouseworkAudit(houseworkAudit{
		Action: "Done by", By: "@a", At: "18/11/2026 14:28", FromAssignee: "@a", ToAssignee: "@b",
	})
	for _, want := range []string{"Done by: @a", "At: 18/11/2026 14:28", "Assignee: @a -> @b"} {
		if !strings.Contains(rotated, want) {
			t.Errorf("rotated audit %q does not contain %q", rotated, want)
		}
	}

	kept := formatHouseworkAudit(houseworkAudit{
		Action: "Done by", By: "@a", At: "18/11/2026 14:28", FromAssignee: "@a", ToAssignee: "@a",
	})
	if !strings.Contains(kept, "Assignee: @a (no change)") {
		t.Errorf("kept audit %q does not show no change", kept)
	}
}

func TestHandledHouseworkMessages(t *testing.T) {
	handledHouseworkMessages = make(map[string]struct{})

	if isHouseworkMessageHandled("1:2") {
		t.Fatal("new key must not be handled")
	}
	markHouseworkMessageHandled("1:2")
	if !isHouseworkMessageHandled("1:2") {
		t.Fatal("marked key must be handled")
	}

	markHouseworkMessageHandled("")
	if isHouseworkMessageHandled("") {
		t.Fatal("empty key must never be handled")
	}
}
