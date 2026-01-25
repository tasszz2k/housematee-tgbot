package commands

import (
	"fmt"
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
	user := ctx.EffectiveUser
	chat := ctx.EffectiveChat
	logrus.WithFields(logrus.Fields{
		"user_id":   user.Id,
		"username":  user.Username,
		"chat_id":   chat.Id,
		"chat_type": chat.Type,
		"action":    action,
	}).Info(details)
}

// Cancel cancels the conversation.
func Cancel(b *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "cancel", "conversation cancelled")
	_, err := ctx.EffectiveMessage.Reply(
		b, "Operation cancelled.", &gotgbot.SendMessageOpts{
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
		"*Coming Soon*\n\nThis feature is not yet available. Stay tuned for updates!",
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
		"*Access Denied*\n\nThis chat is not authorized to use this command.",
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

func CheckPermission(bot *gotgbot.Bot, ctx *ext.Context) bool {
	var hasPermission bool
	// get channel id from ctx
	channelId := ctx.EffectiveChat.Id
	// check is channel id in the list of allowed channels
	for _, id := range config.GetAppConfig().Telegram.AllowedChannels {
		if id == channelId {
			hasPermission = true
			break
		}
	}

	if !hasPermission {
		logrus.WithFields(logrus.Fields{
			"user_id":   ctx.EffectiveUser.Id,
			"username":  ctx.EffectiveUser.Username,
			"chat_id":   channelId,
			"chat_type": ctx.EffectiveChat.Type,
		}).Warn("permission denied - chat not in allowed list")
		_ = ResponseNotHasPermission(bot, ctx)
	}
	return hasPermission
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
