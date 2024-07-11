package enum

const (
	StartCommand              = "start"
	HelloCommand              = "hello"
	GSheetsCommand            = "gsheets"
	SplitBillCommand          = "splitbill"
	SplitBillAddActionCommand = "splitbill_add"
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
