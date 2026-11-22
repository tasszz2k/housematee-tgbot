package handlers

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"google.golang.org/api/sheets/v4"
	"housematee-tgbot/config"
	"housematee-tgbot/enum"
	"housematee-tgbot/models"
	services "housematee-tgbot/services/gsheets"
	"housematee-tgbot/utilities"
)

func GetHouseworkMap() (houseworkMap map[int]models.Task, err error) {
	// get current sheet info
	svc, spreadsheetId, _, err := GetCurrentSheetInfo()
	if err != nil {
		return
	}

	// Get the number of tasks
	numTasksValue, err := svc.GetValue(context.TODO(), spreadsheetId, config.NumberOfTasksReadRange)
	if err != nil {
		logrus.Errorf("failed to get number of tasks: %s", err.Error())
		return
	}

	numTasks := cast.ToInt(numTasksValue)

	if numTasks == 0 {
		return nil, nil
	}

	// Get the list of tasks
	tasksReadRange := fmt.Sprintf("%s!%s%d:%s%d", config.SeparatedSheetTasksName, config.TaskStartCol, config.TaskStartRow, config.TaskEndCol, config.TaskStartRow+numTasks)
	result, err := svc.Get(context.TODO(), spreadsheetId, tasksReadRange)
	if err != nil {
		logrus.Errorf("failed to get tasks: %s", err.Error())
		return
	}

	// map result to the fixed length array (9 columns: ID, Name, Frequency, LastDone, NextDue, Assignee, TurnsRemaining, ChannelId, Note)
	values := make([][9]string, numTasks)
	for i := 1; i < len(result.Values); i++ {
		for j := 0; j < len(result.Values[i]); j++ {
			if j < 9 {
				values[i-1][j] = cast.ToString(result.Values[i][j])
			}
		}
	}

	// convert the result to a map of tasks with key is the task id
	houseworkMap = make(map[int]models.Task)
	for _, value := range values {
		housework := models.Task{
			ID:             cast.ToInt(value[0]),
			Name:           value[1],
			Frequency:      cast.ToInt(value[2]),
			LastDone:       value[3],
			NextDue:        value[4],
			Assignee:       value[5],
			TurnsRemaining: cast.ToInt(value[6]),
			ChannelId:      cast.ToInt64(value[7]),
			Note:           value[8],
		}
		houseworkMap[housework.ID] = housework
	}
	return houseworkMap, nil
}

func UpdateHousework(svc *services.GSheets, spreadsheetId string, currentSheetName string, housework models.Task, numberOfTask int) error {
	// if the housework has id, update it
	if housework.ID == 0 {
		return fmt.Errorf("housework id is not set")
	}

	// update the housework
	writeRow := config.TaskStartRow + housework.ID
	houseworkWriteRange := fmt.Sprintf("%s!%s%d:%s%d", config.SeparatedSheetTasksName, config.TaskStartCol, writeRow, config.TaskEndCol, writeRow)
	houseworkValues := [][]interface{}{
		{
			housework.ID,
			housework.Name,
			housework.Frequency,
			housework.LastDone,
			housework.NextDue,
			housework.Assignee,
			housework.TurnsRemaining,
			housework.ChannelId,
			housework.Note,
		},
	}
	_, err := svc.Update(context.TODO(), spreadsheetId, houseworkWriteRange, &sheets.ValueRange{
		Values: houseworkValues,
	})

	if err != nil {
		logrus.Errorf("failed to update housework: %s", err.Error())
		return err
	}
	return nil
}

func ConvertHouseworkToMarkdownFormat(housework models.Task) string {
	nextDue := housework.NextDue
	if status := DueStatusText(housework.NextDue); status != "" {
		nextDue = fmt.Sprintf("%s (%s)", housework.NextDue, status)
		if days, err := utilities.DaysUntil(housework.NextDue); err == nil && days <= 0 {
			nextDue = "*" + nextDue + "*"
		}
	}

	note := "-"
	if housework.Note != "" {
		note = fmt.Sprintf("_%s_", housework.Note)
	}

	return fmt.Sprintf(
		"%s *%s*\n"+
			"%s Assignee: %s\n"+
			"%s Turns remaining: %d\n"+
			"%s Every %s\n"+
			"%s Last done: %s\n"+
			"%s Next due: %s\n"+
			"%s Note: %s",
		enum.EmojiBroom, housework.Name,
		enum.EmojiUser, housework.Assignee,
		enum.EmojiHourglass, housework.TurnsRemaining,
		enum.EmojiRepeat, pluralDays(housework.Frequency),
		enum.EmojiCheck, valueOrDash(housework.LastDone),
		enum.EmojiCalendar, valueOrDash(nextDue),
		enum.EmojiNote, note,
	)
}

// DueStatusText describes a due date relative to today, or returns "" if the date cannot be parsed.
func DueStatusText(dateStr string) string {
	days, err := utilities.DaysUntil(dateStr)
	if err != nil {
		return ""
	}
	switch {
	case days == 0:
		return "today"
	case days < 0:
		return "overdue " + pluralDays(-days)
	default:
		return "in " + pluralDays(days)
	}
}

func pluralDays(n int) string {
	if n == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", n)
}

func valueOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// GetTaskWeights reads task weights from Google Sheets columns K:M on the Tasks sheet
// Returns a map[int]map[string]int where the outer key is taskID and inner key is username, value is weight
// Default weight to 1 if not specified or if parsing fails
func GetTaskWeights(svc *services.GSheets, spreadsheetId string) (map[int]map[string]int, error) {
	// Get the count of weight entries
	countValue, err := svc.GetValue(context.TODO(), spreadsheetId, config.TaskWeightsCountCell)
	if err != nil {
		logrus.Errorf("failed to get task weights count: %s", err.Error())
		return nil, err
	}

	count := cast.ToInt(countValue)
	if count == 0 {
		return make(map[int]map[string]int), nil
	}

	// Get the task weights data
	weightsReadRange := fmt.Sprintf("%s!%s%d:%s%d", config.SeparatedSheetTasksName, config.TaskWeightsStartCol, config.TaskWeightsStartRow, config.TaskWeightsEndCol, config.TaskWeightsStartRow+count-1)
	result, err := svc.Get(context.TODO(), spreadsheetId, weightsReadRange)
	if err != nil {
		logrus.Errorf("failed to get task weights: %s", err.Error())
		return nil, err
	}

	// Build the nested map: map[taskID]map[username]weight
	weightsMap := make(map[int]map[string]int)
	for _, row := range result.Values {
		if len(row) < 2 {
			continue
		}

		taskID := cast.ToInt(row[0])
		username := cast.ToString(row[1])
		weight := 1 // default weight

		if len(row) >= 3 {
			parsedWeight := cast.ToInt(row[2])
			if parsedWeight > 0 {
				weight = parsedWeight
			}
		}

		// Initialize inner map if it doesn't exist
		if weightsMap[taskID] == nil {
			weightsMap[taskID] = make(map[string]int)
		}

		weightsMap[taskID][username] = weight
	}

	return weightsMap, nil
}
