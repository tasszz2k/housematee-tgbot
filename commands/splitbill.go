package commands

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"housematee-tgbot/enum"
	"housematee-tgbot/utilities"
	"log"
	"strings"
)

// SplitBill handles the /splitbill command.
func SplitBill(bot *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("/splitbill called")
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
	_, err := ctx.EffectiveMessage.Reply(bot, "Select a bill action:", &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		return fmt.Errorf("failed to send /splitbill response: %w", err)
	}
	return nil
}

func HandleSplitBillActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

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
		err := Todo(bot, ctx)
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

func StartAddSplitBill(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Prompt the user to enter the details
	text := `Please provide the details of the expense in the following format:
[expense name]
[amount]
[date] (default: current date)
[payer] (default: current user)
`
	_, err := ctx.EffectiveMessage.Reply(bot, text, nil)
	if err != nil {
		return err
	}
	return handlers.NextConversationState(enum.AddExpense)
}

func AddExpenseConversationHandler(b *gotgbot.Bot, ctx *ext.Context) error {
	// Parse the user's message and extract the details
	input := strings.Split(ctx.EffectiveMessage.Text, "\n")

	//Add validations here to ensure the message contains all required details
	if len(input) < 2 {
		_, err := ctx.EffectiveMessage.Reply(b, "Please provide all required details in the specified format.", nil)
		return err
	}
	details := make([]string, 4)
	copy(details, input)

	expenseName := details[0]
	amountStr := details[1]
	dateStr := details[2]
	payer := details[3]

	// fulfill default values
	if dateStr == "" {
		dateStr = utilities.GetCurrentDate()
	}
	if payer == "" {
		payer = ctx.EffectiveUser.Username
	}

	// Parse amount
	amount := utilities.ParseAmount(amountStr)

	// TODO: Add new record to Google Sheets and get the ID

	// Reply to user with the details
	response := fmt.Sprintf(`
Status: Success
--- *** ---
ID: 1
Expense name: %s
Amount: %sVND
Date: %s
Payer: @%s
`, expenseName, amount, dateStr, payer)

	_, err := ctx.EffectiveMessage.Reply(b, response, nil)
	if err != nil {
		return err
	}
	return handlers.EndConversation()
}
