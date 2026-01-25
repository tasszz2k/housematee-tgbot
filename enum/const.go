package enum

const (
	StartCommand              = "start"
	HelloCommand              = "hello"
	GSheetsCommand            = "gsheets"
	SplitBillCommand          = "splitbill"
	SplitBillAddActionCommand = "splitbill_add"
	RentCommand               = "rent"
	HouseworkCommand          = "housework"
	SettingsCommand           = "settings"
	FeedbackCommand           = "feedback"
	HelpCommand               = "help"
	CancelCommand             = "cancel"
)

func GetCommandAsText(cmd string) string {
	return "/" + cmd
}

const (
	AddExpense      = "add_expense"
	HouseworkPrefix = "hw"
)

// Rent conversation states
const (
	RentStateTotal    = "rent_state_total"
	RentStateElectric = "rent_state_electric"
	RentStateWater    = "rent_state_water"
	RentStatePayer    = "rent_state_payer"
)

// GSheets action constants
const (
	GSheetsActionPrefix  = "gsheets."
	GSheetsCreateAction  = "create"
	GSheetsConfirmCreate = "confirm_create"
	GSheetsCancelCreate  = "cancel_create"
)
