package commands

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	tgBotHandler "github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"housematee-tgbot/enum"
	"housematee-tgbot/handlers"
	"housematee-tgbot/utilities"
)

// SplitBill handles the /splitbill command.
func SplitBill(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "splitbill", "command called")
	// show buttons for these commands
	// - Supported commands:
	// - /add - Add a new expense to the bill.
	// - /view - show last 5 records as table.
	// - /update - update a record.
	// - /delete - delete a record.
	// - /report - show report.

	// Create an inline keyboard with buttons for each command
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Add", CallbackData: "splitbill.add"},
				{Text: "View", CallbackData: "splitbill.view"},
				{Text: "Update", CallbackData: "splitbill.update"},
				{Text: "Delete", CallbackData: "splitbill.delete"},
			},
			{
				{Text: "Show Report", CallbackData: "splitbill.report"},
			},
		},
	}

	// Reply to the user with the available commands as buttons
	_, err := ctx.EffectiveMessage.Reply(
		bot,
		"*Split Bill*\n\nSelect an action:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: inlineKeyboard,
			ParseMode:   "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send /splitbill response: %w", err)
	}
	return nil
}

func HandleSplitBillActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	logUserAction(ctx, "splitbill_callback", fmt.Sprintf("callback: %s", cb.Data))

	// Check the CallbackData to determine which button was clicked
	switch cb.Data {
	case "splitbill.add":
		// Handle the /splitbill button click
		err := StartAddSplitBill(bot, ctx)
		if err != nil {
			return err
		}
	case "splitbill.view":
		// Handle the /housework button click
		err := HandleSplitBillViewActionCallback(bot, ctx)
		if err != nil {
			return err
		}
	case "splitbill.update":
		// Handle the /gsheets button click
		err := Todo(bot, ctx)
		if err != nil {
			return err
		}
	case "splitbill.delete":
		// Handle the /settings button click
		err := Todo(bot, ctx)
		if err != nil {
			return err
		}
	case "splitbill.report":
		// Handle the /feedback button click
		err := HandleSplitBillReportActionCallback(bot, ctx)
		if err != nil {
			return err
		}
	default:
		// Handle other button clicks (if any)
	}

	// Send a response to acknowledge the button click
	_, err := cb.Answer(
		bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: fmt.Sprintf("You clicked %s", cb.Data),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	return nil
}

func HandleSplitBillViewActionCallback(
	bot *gotgbot.Bot,
	ctx *ext.Context,
) error {
	return handlers.HandleSplitBillViewAction(bot, ctx)
}

func HandleSplitBillReportActionCallback(
	bot *gotgbot.Bot,
	ctx *ext.Context,
) error {
	return handlers.HandleSplitBillReportAction(bot, ctx)
}

func StartAddSplitBill(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "splitbill_add", "starting add expense flow")
	// Prompt the user to enter the details
	htlmText := fmt.Sprintf(
		`Please provide the details of the expense in the following format:
---
[expense name]
[amount]
[date] <i>(auto-filled: %s)</i>
[payer] <i>(auto-filled: @%s)</i>
`, utilities.GetCurrentDate(), ctx.EffectiveUser.Username,
	)
	_, err := ctx.EffectiveMessage.Reply(
		bot, htlmText, &gotgbot.SendMessageOpts{
			ParseMode: "html",
		},
	)
	if err != nil {
		return err
	}
	return tgBotHandler.NextConversationState(enum.AddExpense)
}

func AddExpenseConversationHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	if !CheckPermission(bot, ctx) {
		return nil
	}
	return handlers.HandleExpenseAddAction(bot, ctx)
}
