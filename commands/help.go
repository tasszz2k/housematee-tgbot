package commands

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"housematee-tgbot/enum"
)

func Help(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "help", "command called")

	// Create an inline keyboard with buttons for each command
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: enum.GetCommandAsText(enum.SplitBillCommand), CallbackData: "help.splitbill"},
				{Text: enum.GetCommandAsText(enum.HouseworkCommand), CallbackData: "help.housework"},
				{Text: enum.GetCommandAsText(enum.GSheetsCommand), CallbackData: "help.gsheets"},
			},
			{
				{Text: enum.GetCommandAsText(enum.SettingsCommand), CallbackData: "help.settings"},
				{Text: enum.GetCommandAsText(enum.FeedbackCommand), CallbackData: "help.feedback"},
				{Text: enum.GetCommandAsText(enum.HelloCommand), CallbackData: "help.hello"},
			},
		},
	}

	// Reply to the user with the available commands as buttons
	_, err := ctx.EffectiveMessage.Reply(bot, "Here are the available commands:", &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		return fmt.Errorf("failed to send /help response: %w", err)
	}
	return nil
}

func HandleHelpActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	logUserAction(ctx, "help_callback", fmt.Sprintf("callback: %s", cb.Data))

	// Check the CallbackData to determine which button was clicked
	switch cb.Data {
	case "help.splitbill":
		// Handle the /splitbill button click
		err := SplitBill(bot, ctx)
		if err != nil {
			return err
		}
	case "help.housework":
		// Handle the /housework button click
		err := Housework(bot, ctx)
		if err != nil {
			return err
		}
	case "help.gsheets":
		// Handle the /gsheets button click
		err := GSheets(bot, ctx)
		if err != nil {
			return err
		}
	case "help.settings":
		// Handle the /settings button click
		err := Settings(bot, ctx)
		if err != nil {
			return err
		}
	case "help.feedback":
		// Handle the /feedback button click
		err := Feedback(bot, ctx)
		if err != nil {
			return err
		}
	case "help.hello":
		// Handle the /hello button click
		err := Hello(bot, ctx)
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
