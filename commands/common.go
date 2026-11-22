package commands

import (
	"fmt"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/sirupsen/logrus"

	"housematee-tgbot/config"
	"housematee-tgbot/enum"
)

// logUserAction logs user actions with context (user_id, username, chat_id, chat_type, action)
func logUserAction(ctx *ext.Context, action string, details string) {
	fields := ContextLogFields(ctx)
	fields["action"] = action
	logrus.WithFields(fields).Info(details)
}

// displayUsername returns "@username", or the first name for users without a username.
func displayUsername(user *gotgbot.User) string {
	if user == nil {
		return ""
	}
	if user.Username == "" {
		return user.FirstName
	}
	return "@" + user.Username
}

// Cancel cancels the conversation.
func Cancel(b *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "cancel", "conversation cancelled")
	_, err := ctx.EffectiveMessage.Reply(
		b, enum.EmojiCross+" <b>Cancelled</b>\n\nThe current operation is stopped.", &gotgbot.SendMessageOpts{
			ParseMode: "html",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send cancel message: %w", err)
	}
	return handlers.EndConversation()
}

// Todo is a simple command that replies to the user with a not implemented message.
func Todo(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "todo", "action not implemented")
	_, err := ctx.EffectiveMessage.Reply(
		bot,
		enum.EmojiSoon+" *Coming Soon*\n\nThis feature is not available yet.",
		&gotgbot.SendMessageOpts{
			ParseMode: "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send todo message: %w", err)
	}
	return nil
}

// NoCommands Create a matcher which only matches text, which is not a command.
func NoCommands(msg *gotgbot.Message) bool {
	return message.Text(msg) && !message.Command(msg)
}

func ResponseNotHasPermission(bot *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(
		bot,
		enum.EmojiLock+" *Access Denied*\n\nThis chat cannot use this command.",
		&gotgbot.SendMessageOpts{
			ParseMode: "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send permission denied message: %w", err)
	}
	return nil
}

// HandleCommands add middleware to check if a user has permission to use the command
func HandleCommands(bot *gotgbot.Bot, ctx *ext.Context) error {
	// get command from the context
	command := getCommandFromMessage(bot, ctx.Message)
	switch command {
	case enum.HelloCommand, enum.StartCommand:
		return Hello(bot, ctx)
	case enum.GSheetsCommand:
		if !CheckPermission(bot, ctx) {
			return nil
		}
		return GSheets(bot, ctx)
	case enum.SplitBillCommand:
		if !CheckPermission(bot, ctx) {
			return nil
		}
		return SplitBill(bot, ctx)
	case enum.SplitBillAddActionCommand:
		if !CheckPermission(bot, ctx) {
			return nil
		}
		return StartAddSplitBill(bot, ctx)
	// Note: RentCommand is handled by the conversation handler in main.go, not here
	case enum.HouseworkCommand:
		if !CheckPermission(bot, ctx) {
			return nil
		}
		return Housework(bot, ctx)
	case enum.SettingsCommand:
		if !CheckPermission(bot, ctx) {
			return nil
		}
		return Settings(bot, ctx)
	case enum.FeedbackCommand:
		return Feedback(bot, ctx)
	case enum.HelpCommand:
		if !CheckPermission(bot, ctx) {
			return nil
		}
		return Help(bot, ctx)
	case enum.CancelCommand:
		return Cancel(bot, ctx)
	}

	// check some shortcut commands
	if strings.HasPrefix(command, enum.HouseworkPrefix) {
		if !CheckPermission(bot, ctx) {
			return nil
		}
		return MarkAsDoneHouseworkByShortcut(bot, ctx)
	}

	return nil
}

// CheckPermission reports whether the chat and the sender may use protected
// features. On denial it tells the sender: an alert for button clicks, a reply
// for messages.
func CheckPermission(bot *gotgbot.Bot, ctx *ext.Context) bool {
	telegramConfig := config.GetAppConfig().Telegram
	if isPermitted(ctx.EffectiveChat, ctx.EffectiveUser, telegramConfig.AllowedChannels, telegramConfig.AllowedUsers) {
		return true
	}

	logrus.WithFields(ContextLogFields(ctx)).Warn("permission denied - chat or user not in allowed list")
	if cb := ctx.CallbackQuery; cb != nil {
		_, _ = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "Access denied. This chat or user cannot use this bot.",
			ShowAlert: true,
		})
		return false
	}
	if ctx.EffectiveMessage != nil {
		_ = ResponseNotHasPermission(bot, ctx)
	}
	return false
}

// RequirePermission guards handlers that Telegram can trigger without a command,
// such as button callbacks and conversation entry points. Callback data comes
// from the client, and buttons survive forwarding to other chats, so every
// callback must check the chat again.
func RequirePermission(next func(*gotgbot.Bot, *ext.Context) error) func(*gotgbot.Bot, *ext.Context) error {
	return func(bot *gotgbot.Bot, ctx *ext.Context) error {
		if !CheckPermission(bot, ctx) {
			return nil
		}
		return next(bot, ctx)
	}
}

// IsAllowedChat reports whether the chat is in the allowed list.
func IsAllowedChat(chatID int64) bool {
	return slices.Contains(config.GetAppConfig().Telegram.AllowedChannels, chatID)
}

// isPermitted denies inline-mode callbacks (no chat) by design. An empty
// allowedUsers list keeps the old behavior: every member of an allowed chat
// has access.
func isPermitted(chat *gotgbot.Chat, user *gotgbot.User, allowedChats, allowedUsers []int64) bool {
	if chat == nil || !slices.Contains(allowedChats, chat.Id) {
		return false
	}
	if len(allowedUsers) == 0 {
		return true
	}
	return user != nil && slices.Contains(allowedUsers, user.Id)
}

// ContextLogFields returns user and chat log fields. It is safe when the update
// has no user or no chat.
func ContextLogFields(ctx *ext.Context) logrus.Fields {
	fields := logrus.Fields{}
	if user := ctx.EffectiveUser; user != nil {
		fields["user_id"] = user.Id
		fields["username"] = user.Username
	}
	if chat := ctx.EffectiveChat; chat != nil {
		fields["chat_id"] = chat.Id
		fields["chat_type"] = chat.Type
	}
	return fields
}

func getCommandFromMessage(b *gotgbot.Bot, msg *gotgbot.Message) string {
	text := msg.Text
	if msg.Caption != "" {
		text = msg.Caption
	}

	var cmd string
	triggers := []rune{
		'/',
	}
	for _, t := range triggers {
		if r, _ := utf8.DecodeRuneInString(text); r != t {
			continue
		}

		split := strings.Split(strings.ToLower(strings.Fields(text)[0]), "@")
		if len(split) > 1 && split[1] != strings.ToLower(b.User.Username) {
			return ""
		}
		cmd = split[0][1:]
		break
	}
	if cmd == "" {
		return ""
	}

	if len(msg.Entities) != 0 && msg.Entities[0].Offset == 0 && msg.Entities[0].Type != "bot_command" {
		return ""
	}

	return cmd
}
