package models

type Task struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	Frequency      int    `json:"frequency"`
	LastDone       string `json:"last_done"`
	NextDue        string `json:"next_due"`
	Assignee       string `json:"assignee"`
	TurnsRemaining int    `json:"turns_remaining"` // Number of turns remaining before rotation
	ChannelId      int64  `json:"channel_id"`
	Note           string `json:"note"`
}

// TaskWeight defines the rotation weight for a member on a specific task
type TaskWeight struct {
	TaskID   int    `json:"task_id"`
	Username string `json:"username"`
	Weight   int    `json:"weight"` // Number of consecutive turns this member should complete
}
