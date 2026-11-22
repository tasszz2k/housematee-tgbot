package commands

import (
	"testing"

	"github.com/PaulSonOfLars/gotgbot/v2"
)

func TestIsPermitted(t *testing.T) {
	const allowedChat, otherChat int64 = -100, -200
	const housemate, stranger int64 = 1, 2
	chats := []int64{allowedChat}
	users := []int64{housemate}

	testCases := []struct {
		name         string
		chat         *gotgbot.Chat
		user         *gotgbot.User
		allowedUsers []int64
		expected     bool
	}{
		{name: "allowed chat, no user list", chat: &gotgbot.Chat{Id: allowedChat}, user: &gotgbot.User{Id: stranger}, expected: true},
		{name: "other chat", chat: &gotgbot.Chat{Id: otherChat}, user: &gotgbot.User{Id: housemate}, expected: false},
		{name: "nil chat (inline message)", chat: nil, user: &gotgbot.User{Id: housemate}, expected: false},
		{name: "allowed chat, listed user", chat: &gotgbot.Chat{Id: allowedChat}, user: &gotgbot.User{Id: housemate}, allowedUsers: users, expected: true},
		{name: "allowed chat, unlisted user", chat: &gotgbot.Chat{Id: allowedChat}, user: &gotgbot.User{Id: stranger}, allowedUsers: users, expected: false},
		{name: "allowed chat, nil user with user list", chat: &gotgbot.Chat{Id: allowedChat}, user: nil, allowedUsers: users, expected: false},
		{name: "other chat, listed user", chat: &gotgbot.Chat{Id: otherChat}, user: &gotgbot.User{Id: housemate}, allowedUsers: users, expected: false},
	}

	for _, tc := range testCases {
		if actual := isPermitted(tc.chat, tc.user, chats, tc.allowedUsers); actual != tc.expected {
			t.Errorf("%s: expected %v, got %v", tc.name, tc.expected, actual)
		}
	}
}
