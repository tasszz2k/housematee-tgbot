package models

type Expense struct {
	ID           uint32   `json:"id"`
	Name         string   `json:"name"`
	Amount       string   `json:"amount"`
	Date         string   `json:"date"`
	Payer        string   `json:"payer"`
	Participants []string `json:"participants"`
	Note         string   `json:"note"`
}
