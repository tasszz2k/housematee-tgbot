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

	// Report sheet - updated for new template with expanded rent section
	ReportStartCell  = "I3"
	ReportEndCell    = "M9"  // Rows 3-9: Header, Expenses, Electric, Water, Other Fees, Total Rent, Total
	BalanceStartRow  = 13    // Data starts at row 13 (row 11 is label, row 12 is header)
	BalanceStartCell = "I13" // Data starts at I13
	BalanceEndCol    = "M"   // BalanceEndRow = BalanceStartRow + numberOfMembers - 1

	// Rent section cells - bot writes Amount column (J) and Payer (M8)
	RentElectricCell  = "J5" // Electric amount
	RentWaterCell     = "J6" // Water amount
	RentOtherFeesCell = "J7" // Other fees amount
	RentTotalCell     = "J8" // Total rent amount
	RentPayerCell     = "M8" // Payer username

	// Tasks sheet
	SeparatedSheetTasksName = "Tasks"
	TaskStartRow            = 2
	TaskStartCol            = "A"
	TaskEndCol              = "H" // A-H: ID, Name, Frequency, LastDone, NextDue, Assignee, ChannelId, Note
	NumberOfTasksCell       = "B1"
	NumberOfTasksReadRange  = "Tasks!B1"

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
