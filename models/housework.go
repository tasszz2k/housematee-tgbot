package models

type Task struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Frequency int    `json:"frequency"`
	LastDone  string `json:"last_done"`
	NextDue   string `json:"next_due"`
	Assignee  string `json:"assignee"`
	ChannelId int64  `json:"channel_id"`
	Note      string `json:"note"`
}
