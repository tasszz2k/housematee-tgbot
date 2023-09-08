package handlers

import (
	"context"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	tgBotHandler "github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"google.golang.org/api/sheets/v4"
	"housematee-tgbot/config"
	"housematee-tgbot/models"
	services "housematee-tgbot/services/gsheets"
	"housematee-tgbot/utilities"
	"strings"
)

func HandleSplitBillViewAction(bot *gotgbot.Bot, ctx *ext.Context) error {
	readRange, err := getLast5ExpenseReadRange()
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, err.Error(), nil)
		return err
	} else if readRange == "" { // no data found
		_, err := ctx.EffectiveMessage.Reply(bot, "No data found.", nil)
		return err
	}

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
	markdownList := "*Here are the last 5 expenses:*\n---\n"

	// Iterate over the expenses and format them as list items
	for _, expense := range expenses {
		formattedExpense := convertExpenseModelToMarkdown(expense)

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

func convertExpenseModelToMarkdown(expense models.Expense) string {
	participants := "*everyone*"
	if len(expense.Participants) > 0 {
		participants = strings.Join(
			expense.Participants,
			", ",
		) // Join participant names with commas
	}

	// Format the expense data with bold keys
	note := fmt.Sprintf("_%s_", expense.Note)
	formattedExpense := fmt.Sprintf(
		"• *ID*: %d\n  *Name*: %s\n  *Amount*: %s\n  *Date*: %s\n  *Payer*: %s\n  *Participants*: %s\n  *Note*: %s\n\n",
		expense.ID,
		expense.Name,
		expense.Amount,
		expense.Date,
		expense.Payer,
		participants,
		note,
	)
	return formattedExpense
}

func getLast5ExpenseReadRange() (string, error) {
	// read spreadsheetId from config
	svc := services.GetGSheetsSvc()
	spreadsheetId := config.GetAppConfig().GoogleSheets.SpreadsheetId
	currentSheetName, err := svc.GetValue(
		context.TODO(),
		spreadsheetId,
		config.CurrentSheetNameIndex,
	)
	if err != nil {
		logrus.Errorf("failed to get current sheet name: %s", err.Error())
		return "", err
	}

	nextExpenseIdValue, err := getNextExpenseId()
	if err != nil {
		return "", err
	}

	nextExpenseId := cast.ToInt(nextExpenseIdValue)
	// get last 5 expenses read range
	if nextExpenseId == 1 {
		return "", nil
	}

	// example:
	// currentSheetName = "9/2023"
	// nextExpenseId = 7 => currentExpenseId = 6
	// ExpensesStartRow = 3
	// lastExpenseRow = 9
	// => return "9/2023!A5:G9"

	lastExpenseId := nextExpenseId - 1
	lastExpenseRow := config.ExpensesStartRow + lastExpenseId

	readRangeStartRow := lastExpenseRow - 4 // (-5+1)
	readRangeEndRow := lastExpenseRow

	if nextExpenseId <= 5 {
		// example:
		// currentSheetName = "9/2023"
		// nextExpenseId = 4
		// ExpensesStartRow = 3
		// => return "9/2023!A4:G6"

		readRangeStartRow = config.ExpensesStartRow + 1
	}

	readRange := fmt.Sprintf(
		"%s!%s%d:%s%d",
		currentSheetName,
		config.ExpensesStartCol,
		readRangeStartRow,
		config.ExpensesEndCol,
		readRangeEndRow,
	)
	return readRange, nil
}

// HandleExpenseAddAction handles the /splitbill.add command.
// Get the expense details from the user and add a new record to Google Sheets.
// Update next expense ID in Google Sheets.
func HandleExpenseAddAction(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Parse the user's message and extract the details
	input := strings.Split(ctx.EffectiveMessage.Text, "\n")

	//Add validations here to ensure the message contains all required details
	if len(input) < 2 {
		_, err := ctx.EffectiveMessage.Reply(
			bot,
			"Please provide at least the expense name and amount.",
			nil,
		)
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
		payer = "@" + ctx.EffectiveUser.Username
	}

	// Parse amount
	amount := utilities.ParseAmount(amountStr)

	// validate input
	if err := checkValidExpenseInput(expenseName, amount, dateStr, payer); err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, err.Error(), nil)
		return err
	}

	expense := models.Expense{
		Name:         expenseName,
		Amount:       amount,
		Date:         dateStr,
		Payer:        payer,
		Participants: []string{},
		Note:         "added from telegram bot",
	}

	// TODO: Add new record to Google Sheets and get the ID
	newExpense, err := addNewExpense(expense)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, err.Error(), nil)
		return err
	}

	// Reply to user with the details
	response := convertExpenseModelToMarkdown(*newExpense)
	_, err = ctx.EffectiveMessage.Reply(bot, response, &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	})
	if err != nil {
		return err
	}
	return tgBotHandler.EndConversation()
}

func addNewExpense(expense models.Expense) (*models.Expense, error) {
	// read spreadsheetId from config
	svc, spreadsheetId, currentSheetName, err := GetCurrentSheetInfo()
	if err != nil {
		return nil, err
	}

	// get next expense id
	nextExpenseId, err := getNextExpenseId()
	if err != nil {
		return nil, err
	}

	// get next expense row
	nextExpenseRow := config.ExpensesStartRow + nextExpenseId

	// get next expense range
	expenseRange := fmt.Sprintf(
		"%s!%s%d:%s%d",
		currentSheetName,
		config.ExpensesStartCol,
		nextExpenseRow,
		config.ExpensesEndCol,
		nextExpenseRow,
	)

	if expense.Participants == nil {
		expense.Participants = []string{}
	}
	// write expense to Google Sheets
	expense.ID = cast.ToUint32(nextExpenseId)
	expenseValues := [][]interface{}{
		{
			expense.ID,
			expense.Name,
			cast.ToInt(expense.Amount),
			expense.Date,
			expense.Payer,
			strings.Join(expense.Participants, ","),
			expense.Note,
		},
	}
	_, err = svc.Update(context.TODO(), spreadsheetId, expenseRange, &sheets.ValueRange{
		Values: expenseValues,
	})
	if err != nil {
		logrus.Errorf("failed to update expense: %s", err.Error())
		return nil, err
	}

	expense.Amount = utilities.FormatMoney(cast.ToInt(expense.Amount))

	// update next expense id
	nextExpenseId = nextExpenseId + 1
	if _, err := svc.Update(context.TODO(), spreadsheetId, config.NextExpenseIdIndex, &sheets.ValueRange{
		Values: [][]interface{}{
			{nextExpenseId},
		},
	}); err != nil {
		logrus.Errorf("failed to update next expense id: %s", err.Error())
		return nil, err
	}

	// return new expense

	return &expense, nil
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

func getNextExpenseId() (int, error) {
	// read spreadsheetId from config
	svc, spreadsheetId, _, err := GetCurrentSheetInfo()
	if err != nil {
		return 0, err
	}

	// get next expense id
	nextExpenseIdValue, err := svc.GetValue(
		context.TODO(),
		spreadsheetId,
		config.NextExpenseIdIndex,
	)
	if err != nil {
		logrus.Errorf("failed to get next expense id: %s", err.Error())
		return 0, err
	}
	nextExpenseId := cast.ToInt(nextExpenseIdValue)
	return nextExpenseId, nil
}
