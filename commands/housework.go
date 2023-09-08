package commands

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"log"
)

// Housework handles the /housework command.
func Housework(bot *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("/housework called")
	// show buttons for these commands
	// - Supported commands:
	// - /list - List all housework.
	// - /add - Add new housework.
	// - /update - update a record.
	// - /delete - delete a record.

	// Create an inline keyboard with buttons for each command
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "List", CallbackData: "housework.list"},
				{Text: "Add", CallbackData: "housework.add"},
				{Text: "Update", CallbackData: "housework.update"},
				{Text: "Delete", CallbackData: "housework.delete"},
			},
		},
	}

	// Reply to the user with the available commands as buttons
	_, err := ctx.EffectiveMessage.Reply(bot, "Select a housework action:", &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}
	return nil
}

func HandleHouseworkActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	// Check the CallbackData to determine which button was clicked
	switch cb.Data {
	case "housework.list":
		// Handle the /splitbill button click
		err := Todo(bot, ctx)
		if err != nil {
			return err
		}
	case "housework.add":
		// Handle the /housework button click
		err := Todo(bot, ctx)
		if err != nil {
			return err
		}
	case "housework.update":
		// Handle the /gsheets button click
		err := Todo(bot, ctx)
		if err != nil {
			return err
		}
	case "housework.delete":
		// Handle the /settings button click
		err := Todo(bot, ctx)
		if err != nil {
			return err
		}
	default:
		// Handle other button clicks (if any)
	}

	// Send a response to acknowledge the button click
	_, err := cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
		Text: fmt.Sprintf("You clicked %s", cb.Data),
	})
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	return nil
}
