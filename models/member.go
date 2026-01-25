package models

type Member struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Weight   int    `json:"weight"`
}
