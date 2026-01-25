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
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("*Error*\n\n%s", err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	} else if readRange == "" { // no data found
		_, err := ctx.EffectiveMessage.Reply(bot, "*No Expenses*\n\nNo expenses recorded yet. Use *Add* to create your first expense.", &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	spreadsheetId := config.GetAppConfig().GoogleSheets.SpreadsheetId
	svc := services.GetGSheetsSvc()
	resp, err := svc.Get(context.TODO(), spreadsheetId, readRange)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("*Error*\n\n%s", err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	// convert resp to models.Expense array
	if len(resp.Values) == 0 {
		_, err := ctx.EffectiveMessage.Reply(bot, "*No Expenses*\n\nNo expenses recorded yet. Use *Add* to create your first expense.", &gotgbot.SendMessageOpts{ParseMode: "markdown"})
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
	markdownList := "*Recent Expenses*\n\n"

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
		config.CurrentSheetNameCell,
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
	lastExpenseRow := config.ExpenseStartRow + lastExpenseId

	readRangeStartRow := lastExpenseRow - 4 // (-5+1)
	readRangeEndRow := lastExpenseRow

	if nextExpenseId <= 5 {
		// example:
		// currentSheetName = "9/2023"
		// nextExpenseId = 4
		// ExpensesStartRow = 3
		// => return "9/2023!A4:G6"

		readRangeStartRow = config.ExpenseStartRow + 1
	}

	readRange := fmt.Sprintf(
		"%s!%s%d:%s%d",
		currentSheetName,
		config.ExpenseStartCol,
		readRangeStartRow,
		config.ExpenseEndCol,
		readRangeEndRow,
	)
	return readRange, nil
}

// HandleExpenseAddAction handles the /splitbill.add command.
// Get the expense details from the user and add a new record to Google Sheets.
// Update next expense ID in Google Sheets.
func HandleExpenseAddAction(bot *gotgbot.Bot, ctx *ext.Context) (err error) {
	// Parse the user's message and extract the details
	input := strings.Split(ctx.EffectiveMessage.Text, "\n")

	//Add validations here to ensure the message contains all required details
	if len(input) < 2 {
		_, err := ctx.EffectiveMessage.Reply(
			bot,
			"*Invalid Input*\n\nPlease provide at least the expense name and amount.",
			&gotgbot.SendMessageOpts{ParseMode: "markdown"},
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
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("*Validation Error*\n\n%s", err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
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

	var newExpense *models.Expense
	// Check if this is a rent expense - redirect to /rent command
	isRentExpense := strings.ToLower(strings.TrimSpace(expenseName)) == config.ExpenseNameRent
	if isRentExpense {
		// Redirect user to use the new /rent command for detailed breakdown
		_, err := ctx.EffectiveMessage.Reply(
			bot,
			"*Rent Detected*\n\nTo add rent with detailed breakdown (electric, water, other fees), please use the /rent command.",
			&gotgbot.SendMessageOpts{ParseMode: "markdown"},
		)
		if err != nil {
			return err
		}
		return tgBotHandler.EndConversation()
	}

	// Add regular expense to Google Sheets
	newExpense, err = addNewExpense(expense)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("*Failed to Add Expense*\n\n%s", err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	// Reply to user with the details
	response := "*Expense Added*\n\n" + convertExpenseModelToMarkdown(*newExpense)
	_, err = ctx.EffectiveMessage.Reply(bot, response, &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	})
	if err != nil {
		return err
	}
	return tgBotHandler.EndConversation()
}

// upsertRentExpense is deprecated - use /rent command with handlers/rent.go instead

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
	nextExpenseRow := config.ExpenseStartRow + nextExpenseId

	// get next expense range
	expenseRange := fmt.Sprintf(
		"%s!%s%d:%s%d",
		currentSheetName,
		config.ExpenseStartCol,
		nextExpenseRow,
		config.ExpenseEndCol,
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
	nextExpenseIdCell := config.GetNextExpenseIdCell(currentSheetName)
	if _, err := svc.Update(context.TODO(), spreadsheetId, nextExpenseIdCell, &sheets.ValueRange{
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
	svc, spreadsheetId, currentSheetName, err := GetCurrentSheetInfo()
	if err != nil {
		return 0, err
	}

	// get next expense id
	nextExpenseIdCell := config.GetNextExpenseIdCell(currentSheetName)
	nextExpenseIdValue, err := svc.GetValue(
		context.TODO(),
		spreadsheetId,
		nextExpenseIdCell,
	)
	if err != nil {
		logrus.Errorf("failed to get next expense id: %s", err.Error())
		return 0, err
	}
	nextExpenseId := cast.ToInt(nextExpenseIdValue)
	return nextExpenseId, nil
}

// HandleSplitBillReportAction handles the /splitbill.report command.
// Sample data format:
// Report
// =================================================================
//
//	Amount	    Average	     Note
//
// Expenses	1,999,500 ₫	999,750 ₫
// Rent	    5,500,000 ₫	2,750,000 ₫	 @tasszz2k
// Total	7,499,500 ₫	3,749,750 ₫
//
// Balances
// Username	    Total Paid	 Have to pay  Balance     Final Balance
// @tasszz2k	1,720,900 ₫	 999,750 ₫	  721,150 ₫	  3,471,150 ₫
// @ng0cth1nh	278,600 ₫	 999,750 ₫	  -721,150 ₫  -3,471,150 ₫
// =================================================================
func HandleSplitBillReportAction(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Read the spreadsheet data and calculate the report
	report, err := generateSplitBillReport()
	if err != nil {
		return err
	}

	// Send the report to the user
	_, err = ctx.EffectiveMessage.Reply(bot, report, &gotgbot.SendMessageOpts{
		ParseMode: "Markdown",
	})
	return err
}

func generateSplitBillReport() (result string, err error) {
	// read spreadsheetId from config
	svc, spreadsheetId, currentSheetName, err := GetCurrentSheetInfo()
	if err != nil {
		return "", err
	}

	report, err := getReport(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return "", err
	}

	balances, err := getBalances(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return "", err
	}

	return renderReportMarkdown(report, balances), nil
}

func getBalances(svc *services.GSheets, spreadsheetId string, currentSheetName string) (models.Balance, error) {
	// get balances read range
	numberOfMembers, err := GetNumberOfMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return models.Balance{}, err
	}
	balancesReadRange := getBalancesReadRange(currentSheetName, numberOfMembers)

	// get balances data
	balancesData, err := svc.Get(context.TODO(), spreadsheetId, balancesReadRange)
	if err != nil {
		logrus.Errorf("failed to get balances data: %s", err.Error())
		return models.Balance{}, err
	}

	// convert balances data to models.Balance
	balances := convertBalancesDataToBalanceModel(balancesData.Values)
	return balances, nil
}

func convertBalancesDataToBalanceModel(values [][]interface{}) (balances models.Balance) {
	// convert data to [numberOfMembers][5] string array (no header row)
	balancesArray := make([][5]string, len(values))
	for i, row := range values {
		for j, col := range row {
			if j < 5 {
				balancesArray[i][j] = cast.ToString(col)
			}
		}
	}

	balances.Users = make(map[string]models.BalanceData)
	// process all rows (no header to skip)
	for i := 0; i < len(balancesArray); i++ {
		username := balancesArray[i][0]
		if username == "" {
			continue
		}
		balances.Users[username] = models.BalanceData{
			TotalPaid:    balancesArray[i][1],
			HaveToPay:    balancesArray[i][2],
			Balance:      balancesArray[i][3],
			FinalBalance: balancesArray[i][4],
		}
	}

	return balances
}

func renderReportMarkdown(report models.Report, balances models.Balance) string {
	text := "\U0001F4CA *Report*\n\n"

	text += "\U0001F6D2 *Expenses*\n"
	text += "\u2022 *Amount*: " + report.Expenses.Amount + "\n\n"

	text += "\U0001F3E0 *Rent*\n"
	if report.Rent.Amount == "" || report.Rent.Amount == "0" {
		text += "\u2022 *Amount*: _not paid_\n"
	} else {
		text += "\u2022 *Amount*: " + report.Rent.Amount + "\n"
		if report.Rent.Note != "" && report.Rent.Note != "x" {
			text += "\u2022 *Payer*: _" + report.Rent.Note + "_\n"
		}
	}
	text += "\n"

	text += "\U0001F4B0 *Total*\n"
	text += "\u2022 *Amount*: " + report.Total.Amount + "\n\n"

	text += "-----\n\U0001F4B3 *Balances*\n\n"

	balanceMap := balances.Users
	for username, balance := range balanceMap {
		text += "\U0001F464 *" + username + "*:\n"
		text += "\u2022 *Total Paid*: " + balance.TotalPaid + "\n"
		text += "\u2022 *Expense Balance*: " + balance.HaveToPay + "\n"
		text += "\u2022 *Rent Balance*: " + balance.Balance + "\n"
		text += "\u2022 *Final Balance*: " + balance.FinalBalance + "\n\n"
	}

	return text
}

func getReport(svc *services.GSheets, spreadsheetId string, currentSheetName string) (models.Report, error) {
	// get report read range
	reportReadRange := getReportReadRange(currentSheetName)

	// get report data
	reportData, err := svc.Get(context.TODO(), spreadsheetId, reportReadRange)
	if err != nil {
		logrus.Errorf("failed to get report data: %s", err.Error())
		return models.Report{}, err
	}

	// convert report data to models.Report
	report := convertReportDataToReportModel(reportData.Values)
	return report, nil
}

func convertReportDataToReportModel(data [][]any) models.Report {
	// New template format (7 rows):
	// Row 0: Header (Category, Amount, @tasszz2k, @ng0cth1nh, Payer)
	// Row 1: Expenses
	// Row 2: Electric
	// Row 3: Water
	// Row 4: Other Fees
	// Row 5: Total Rent
	// Row 6: Total

	// convert data to [7][5] string array
	reportArray := make([][5]string, 7)
	for i, row := range data {
		if i >= 7 {
			break
		}
		for j, col := range row {
			if j >= 5 {
				break
			}
			reportArray[i][j] = cast.ToString(col)
		}
	}

	return models.Report{
		Expenses: models.ReportData{
			Amount:  reportArray[1][1],
			Average: reportArray[1][2],
			Note:    reportArray[1][4], // Payer column
		},
		Rent: models.ReportData{
			Amount:  reportArray[5][1], // Total Rent row
			Average: reportArray[5][2],
			Note:    reportArray[5][4], // Payer column
		},
		Total: models.ReportData{
			Amount:  reportArray[6][1], // Total row
			Average: reportArray[6][2],
			Note:    reportArray[6][4],
		},
	}
}

func getReportReadRange(currentSheetName string) string {
	return fmt.Sprintf("%s!%s:%s", currentSheetName, config.ReportStartCell, config.ReportEndCell)
}

func getBalancesReadRange(currentSheetName string, numberOfMembers int) string {
	return fmt.Sprintf(
		"%s!%s:%s%d",
		currentSheetName,
		config.BalanceStartCell,
		config.BalanceEndCol,
		config.BalanceStartRow+numberOfMembers-1,
	)
}
