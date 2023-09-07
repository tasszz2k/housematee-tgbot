package commands

import (
	"context"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/spf13/cast"
	"housematee-tgbot/config"
	"housematee-tgbot/enum"
	"housematee-tgbot/models"
	services "housematee-tgbot/services/gsheets"
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
	_, err := ctx.EffectiveMessage.Reply(bot, "Select a housework action:", &gotgbot.SendMessageOpts{
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
	htlmText := fmt.Sprintf(`Please provide the details of the expense in the following format:
---
[expense name]
[amount]
[date] <i>(auto-filled: %s)</i>
[payer] <i>(auto-filled: @%s)</i>
`, utilities.GetCurrentDate(), ctx.EffectiveUser.Username)
	_, err := ctx.EffectiveMessage.Reply(bot, htlmText, &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
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
		_, err := ctx.EffectiveMessage.Reply(b, "Please provide at least the expense name and amount.", nil)
		return err
	}
	details := make([]string, 4)
	copy(details, input)

	expenseName := details[0]
	amountStr := details[1]
	dateStr := details[2]
	payer := details[3]

	// fulfill default values
	if expenseName == "" {
		expenseName = fmt.Sprintf("Expense %s - %s", ctx.EffectiveUser.Username, utilities.GetCurrentDate())
	}
	if dateStr == "" {
		dateStr = utilities.GetCurrentDate()
	}
	if payer == "" {
		payer = ctx.EffectiveUser.Username
	}

	// Parse amount
	amount := utilities.ParseAmount(amountStr)

	// validate input
	if err := checkValidExpenseInput(expenseName, amount, dateStr, payer); err != nil {
		_, err := ctx.EffectiveMessage.Reply(b, err.Error(), nil)
		return err
	}

	// TODO: Add new record to Google Sheets and get the ID

	// Reply to user with the details
	response := fmt.Sprintf(`
Status: Success
--- *** ---
ID: 1
Expense name: %s
Amount: %s ₫
Date: %s
Payer: @%s
`, expenseName, amount, dateStr, payer)

	_, err := ctx.EffectiveMessage.Reply(b, response, nil)
	if err != nil {
		return err
	}
	return handlers.EndConversation()
}

func checkValidExpenseInput(name string, amount string, str string, payer string) error {
	if name == "" {
		return fmt.Errorf("expense name cannot be empty")
	}
	if amount == "" {
		return fmt.Errorf("amount cannot be empty")
	} else if !utilities.IsNumeric(amount) {
		return fmt.Errorf("cannot parse amount to number")
	}
	if str == "" {
		return fmt.Errorf("date cannot be empty")
	}
	if payer == "" {
		return fmt.Errorf("payer cannot be empty")
	}
	return nil
}

func HandleSplitBillViewActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	readRange := "9/2023!A4:G7"
	spreadsheetId := config.GetAppConfig().GoogleSheets.SpreadsheetId
	svc := services.GetGSheetsSvc()
	resp, err := svc.Get(context.TODO(), spreadsheetId, readRange)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, err.Error(), nil)
		return err
	}

	// convert resp to models.Expense array
	if len(resp.Values) == 0 {
		_, err := ctx.EffectiveMessage.Reply(bot, "No data found.", nil)
		return err
	}

	expenses := make([]models.Expense, 0, len(resp.Values))
	respValues := make([][7]interface{}, len(resp.Values))
	// copy resp.Values to respValue
	for i, row := range resp.Values {
		for j, col := range row {
			respValues[i][j] = col
		}
	}

	for _, row := range respValues {

		// map row to Expense
		expense := models.Expense{
			ID:           cast.ToUint32(row[0]),
			Name:         cast.ToString(row[1]),
			Amount:       cast.ToString(row[2]),
			Date:         cast.ToString(row[3]),
			Payer:        cast.ToString(row[4]),
			Participants: cast.ToStringSlice(row[5]),
			Note:         cast.ToString(row[6]),
		}
		expenses = append(expenses, expense)
	}

	// render html table and response to user
	//                 | ID | Expense name | Amount | Date | Payer |
	//                |:---|:-------------|:-------|:-----|:------|
	//                | 1  | ...          | ...    | ...  | ...   |
	//                | 2  | ...          | ...    | ...  | ...   |

	// Create a Markdown-formatted list with bold keys
	markdownList := ""

	// Iterate over the expenses and format them as list items
	for _, expense := range expenses {
		participants := strings.Join(expense.Participants, ", ") // Join participant names with commas

		// Format the expense data with bold keys
		formattedExpense := fmt.Sprintf(
			"• *ID*: %d\n  *Name*: %s\n  *Amount*: %s\n  *Date*: %s\n  *Payer*: %s\n  *Participants*: %s\n  *Note*: %s\n\n",
			expense.ID, expense.Name, expense.Amount, expense.Date, expense.Payer, participants, expense.Note)

		// Append the formatted expense to the Markdown list
		markdownList += formattedExpense
	}

	// Send the Markdown-formatted list to the user
	_, err = ctx.EffectiveMessage.Reply(bot, markdownList, &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	})
	if err != nil {
		return err
	}

	return nil
}
