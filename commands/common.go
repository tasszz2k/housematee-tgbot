package commands

import (
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"

	"housematee-tgbot/config"
	"housematee-tgbot/enum"
)

// Cancel cancels the conversation.
func Cancel(b *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(
		b, "Oh, goodbye!", &gotgbot.SendMessageOpts{
			ParseMode: "html",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send cancel message: %w", err)
	}
	return handlers.EndConversation()
}

// Todo is a simple command that replies to the user with a hello message.
func Todo(bot *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("/todo called")
	_, err := ctx.EffectiveMessage.Reply(
		bot,
		"This action is not yet implemented in this version of the bot. Stay tuned for future updates!",
		&gotgbot.SendMessageOpts{
			ParseMode: "html",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
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
		"Sorry, you don't have permission to use this command.",
		&gotgbot.SendMessageOpts{
			ParseMode: "html",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
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
