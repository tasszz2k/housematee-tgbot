package models

type Report struct {
	Expenses ReportData
	Rent     ReportData
	Total    ReportData
}

type ReportData struct {
	Amount  string
	Average string
	Note    string
}

type Balance struct {
	Users map[string]BalanceData // map[username]BalanceData
}

type BalanceData struct {
	TotalPaid    string
	HaveToPay    string
	Balance      string
	FinalBalance string // Balance +/- Rent.Average (depending on who pays rent)
}
