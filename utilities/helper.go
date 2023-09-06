package utilities

import "time"

func GetCurrentDate() string {
	return time.Now().Format("02/01/2006")
}
