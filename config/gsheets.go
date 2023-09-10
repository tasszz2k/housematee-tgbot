package config

// Indexes of data in Google Sheets
var (
	// Database sheet
	SeperatedSheetDatabaseName = "Database"
	CurrentSheetNameCell       = "Database!B2"

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

	// Members sheet
	NumberOfMembersCell = "P2"

	// Tasks sheet
	SeparatedSheetTasksName = "Tasks"
	TaskStartRow            = 2
	TaskStartCol            = "A"
	TaskEndCol              = "H"
	NumberOfTasksCell       = "B1"
	NumberOfTasksReadRange  = "Tasks!B1"
)

const (
	ExpenseNameRent = "rent"
)

func GetNextExpenseIdCell(sheetName string) string {
	return sheetName + "!" + NextExpenseIdCell
}
