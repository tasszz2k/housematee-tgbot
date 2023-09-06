package enum

const (
	HelloCommand     = "hello"
	GSheetsCommand   = "gsheets"
	SplitBillCommand = "splitbill"
	HouseworkCommand = "housework"
	SettingsCommand  = "settings"
	FeedbackCommand  = "feedback"
	HelpCommand      = "help"
	CancelCommand    = "cancel"
)

func GetCommandAsText(cmd string) string {
	return "/" + cmd
}

const (
	AddExpense = "add_expense"
)
