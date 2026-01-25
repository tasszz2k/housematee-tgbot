package config

// Indexes of data in Google Sheets
var (
	// Database sheet
	SeperatedSheetDatabaseName = "Database"
	CurrentSheetNameCell       = "Database!B2"

	// Template sheet
	TemplateSheetName = "Template"

	// Expenses sheet
	NextExpenseIdCell = "B2" // example: "9/2023!B2"
	ExpenseStartRow   = 3
	ExpenseStartCol   = "A"
	ExpenseEndCol     = "G"

	ExpenseRentReadRange = "J5:L5"

	// Report sheet
	ReportStartCell  = "I3"
	ReportEndCell    = "M6"
	BalanceStartRow  = 9
	BalanceStartCell = "I9"
	BalanceEndCol    = "M" // BalanceEndRow = BalanceStartRow + numberOfHousemates

	// Tasks sheet
	SeparatedSheetTasksName = "Tasks"
	TaskStartRow            = 2
	TaskStartCol            = "A"
	TaskEndCol              = "I" // Updated: A-I (ID, Name, Frequency, LastDone, NextDue, Assignee, TurnsRemaining, ChannelId, Note)
	NumberOfTasksCell       = "B1"
	NumberOfTasksReadRange  = "Tasks!B1"

	// Task Weights (on same Tasks sheet, columns K-M)
	TaskWeightsCountCell = "Tasks!L1"
	TaskWeightsStartRow  = 3
	TaskWeightsStartCol  = "K"
	TaskWeightsEndCol    = "M"

	// Members sheet
	NumberOfMembersCell = "P2"
	MembersStartRow     = 3
	MembersStartCol     = "O"
	MembersEndCol       = "P"
)

const (
	ExpenseNameRent = "rent"
)

func GetNextExpenseIdCell(sheetName string) string {
	return sheetName + "!" + NextExpenseIdCell
}
