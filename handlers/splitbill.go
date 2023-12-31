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

	var newExpense *models.Expense
	// Add new record to Google Sheets and get the ID
	isRentExpense := strings.ToLower(strings.TrimSpace(expenseName)) == config.ExpenseNameRent
	if isRentExpense {
		// add to rent range. example: "9/2023!J5:L5"
		newExpense, err = upsertRentExpense(expense)
		if err != nil {
			_, err := ctx.EffectiveMessage.Reply(bot, err.Error(), nil)
			return err
		}
	} else {
		newExpense, err = addNewExpense(expense)
		if err != nil {
			_, err := ctx.EffectiveMessage.Reply(bot, err.Error(), nil)
			return err
		}
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

func upsertRentExpense(expense models.Expense) (*models.Expense, error) {
	// read spreadsheetId from config
	svc, spreadsheetId, currentSheetName, err := GetCurrentSheetInfo()
	if err != nil {
		return nil, err
	}

	// get number of members
	numberOfMembers, err := GetNumberOfMembers(svc, spreadsheetId, currentSheetName)

	// get rent range
	rentRange := currentSheetName + "!" + config.ExpenseRentReadRange
	amount := cast.ToInt(expense.Amount)
	average := amount / numberOfMembers
	rentValues := [][]interface{}{
		{
			amount,
			average,
			expense.Payer,
		},
	}

	// write rent to Google Sheets
	_, err = svc.Update(context.TODO(), spreadsheetId, rentRange, &sheets.ValueRange{
		Values: rentValues,
	})
	if err != nil {
		logrus.Errorf("failed to update rent: %s", err.Error())
		return nil, err
	}

	// update rent expense value
	rentNote := "Average: " + utilities.FormatMoney(average)
	return &models.Expense{
		ID:           0,
		Name:         expense.Name,
		Amount:       utilities.FormatMoney(cast.ToInt(expense.Amount)),
		Date:         expense.Date,
		Payer:        expense.Payer,
		Participants: expense.Participants,
		Note:         rentNote,
	}, nil
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
	// convert data to [numberOfMembers+1][5] string array
	balancesArray := make([][5]string, len(values))
	for i, row := range values {
		for j, col := range row {
			balancesArray[i][j] = cast.ToString(col)
		}
	}

	balances.Users = make(map[string]models.BalanceData)
	// skip the first row is header
	for i := 1; i < len(balancesArray); i++ {
		username := balancesArray[i][0]
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
	text := "*Report*\n\n"
	text += "*Expenses*\n"
	text += "• *Amount*: " + report.Expenses.Amount + "\n"
	text += "• *Average*: " + report.Expenses.Average + "\n"
	if report.Expenses.Note != "" {
		text += "• *Note*: _" + report.Expenses.Note + "_\n\n"
	}
	text += "*Rent*\n"
	if report.Rent.Amount == "" {
		text += "• *Amount*: _not paid_\n"
		text += "• *Average*: _not paid_\n"
	} else {
		text += "• *Amount*: " + report.Rent.Amount + "\n"
		text += "• *Average*: " + report.Rent.Average + "\n"
		if report.Rent.Note != "" {
			text += "• *Note*: _" + report.Rent.Note + "_\n\n"
		}
	}
	text += "*Total*\n"
	text += "• *Amount*: " + report.Total.Amount + "\n"
	text += "• *Average*: " + report.Total.Average + "\n"
	if report.Total.Note != "" {
		text += "• *Note*: _" + report.Total.Note + "_\n\n"
	}

	text += "\n-----\n*Balances*\n\n"
	//text += "@tasszz2k:\n"
	//text += "• *Total Paid*: " + report.Balances["@tasszz2k"].TotalPaid + "\n"
	//text += "• *Have to pay*: " + report.Balances["@tasszz2k"].HaveToPay + "\n"
	//text += "• *Balance*: " + report.Balances["@tasszz2k"].Balance + "\n"
	//text += "• *Final Balance*: " + report.Balances["@tasszz2k"].FinalBalance + "\n\n"
	//
	//text += "@ng0cth1nh:\n"
	//text += "• *Total Paid*: " + report.Balances["@ng0cth1nh"].TotalPaid + "\n"
	//text += "• *Have to pay*: " + report.Balances["@ng0cth1nh"].HaveToPay + "\n"
	//text += "• *Balance*: " + report.Balances["@ng0cth1nh"].Balance + "\n"
	//text += "• *Final Balance*: " + report.Balances["@ng0cth1nh"].FinalBalance + "\n"

	balanceMap := balances.Users
	for username, balance := range balanceMap {
		text += "*" + username + "*" + ":\n"
		text += "• *Total Paid*: " + balance.TotalPaid + "\n"
		text += "• *Have to pay*: " + balance.HaveToPay + "\n"
		text += "• *Balance*: " + balance.Balance + "\n"
		text += "• *Final Balance*: " + balance.FinalBalance + "\n\n"
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
	// convert data to [4][4] string array
	reportArray := make([][4]string, 4)
	for i, row := range data {
		for j, col := range row {
			reportArray[i][j] = cast.ToString(col)
		}
	}

	return models.Report{
		Expenses: models.ReportData{
			Amount:  reportArray[1][1],
			Average: reportArray[1][2],
			Note:    reportArray[1][3],
		},
		Rent: models.ReportData{
			Amount:  reportArray[2][1],
			Average: reportArray[2][2],
			Note:    reportArray[2][3],
		},
		Total: models.ReportData{
			Amount:  reportArray[3][1],
			Average: reportArray[3][2],
			Note:    reportArray[3][3],
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
		config.BalanceStartRow+numberOfMembers,
	)
}
