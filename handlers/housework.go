package handlers

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"google.golang.org/api/sheets/v4"
	"housematee-tgbot/config"
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
			values[i-1][j] = cast.ToString(result.Values[i][j])
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
	frequency := fmt.Sprintf("%d days", housework.Frequency)
	note := fmt.Sprintf("_%s_", housework.Note)
	// if the next due is today, add an emoji
	nextDue := housework.NextDue
	if housework.NextDue == utilities.GetCurrentDate() {
		nextDue = fmt.Sprintf("*%s >> Today*", housework.NextDue)
	}

	return fmt.Sprintf(
		"*Name*: %s\n*Frequency*: %s\n*Last done*: %s\n*Next due*: %s\n*Assignee*: %s\n*Turns remaining*: %d\n*Note*: %s",
		housework.Name,
		frequency,
		housework.LastDone,
		nextDue,
		housework.Assignee,
		housework.TurnsRemaining,
		note,
	)
}

// GetTaskWeights reads task weights from the Tasks sheet (columns K-M)
// Returns a slice of TaskWeight for the given taskID, ordered by row order
func GetTaskWeights(taskID int) ([]models.TaskWeight, error) {
	svc, spreadsheetId, _, err := GetCurrentSheetInfo()
	if err != nil {
		return nil, err
	}

	// Get the number of task weights
	numWeightsValue, err := svc.GetValue(context.TODO(), spreadsheetId, config.TaskWeightsCountCell)
	if err != nil {
		logrus.Errorf("failed to get number of task weights: %s", err.Error())
		return nil, err
	}

	numWeights := cast.ToInt(numWeightsValue)
	if numWeights == 0 {
		return nil, nil
	}

	// Read task weights range: Tasks!K3:M{2+numWeights}
	endRow := config.TaskWeightsStartRow + numWeights - 1
	weightsReadRange := fmt.Sprintf("%s!%s%d:%s%d",
		config.SeparatedSheetTasksName,
		config.TaskWeightsStartCol,
		config.TaskWeightsStartRow,
		config.TaskWeightsEndCol,
		endRow,
	)

	result, err := svc.Get(context.TODO(), spreadsheetId, weightsReadRange)
	if err != nil {
		logrus.Errorf("failed to get task weights: %s", err.Error())
		return nil, err
	}

	// Filter weights by taskID and maintain row order
	var weights []models.TaskWeight
	for _, row := range result.Values {
		if len(row) < 3 {
			continue
		}
		rowTaskID := cast.ToInt(row[0])
		if rowTaskID == taskID {
			weight := models.TaskWeight{
				TaskID:   rowTaskID,
				Username: cast.ToString(row[1]),
				Weight:   cast.ToInt(row[2]),
			}
			weights = append(weights, weight)
		}
	}

	return weights, nil
}

// FindNextAssigneeWithWeight finds the next assignee in the rotation based on task weights
// Returns the next assignee username and their weight
func FindNextAssigneeWithWeight(weights []models.TaskWeight, currentAssignee string) (string, int) {
	if len(weights) == 0 {
		return currentAssignee, 1
	}

	// Find current assignee's position in the weights list
	currentIndex := -1
	for i, w := range weights {
		if w.Username == currentAssignee {
			currentIndex = i
			break
		}
	}

	// If current assignee not found, start from the first entry
	if currentIndex == -1 {
		return weights[0].Username, weights[0].Weight
	}

	// Get next assignee (wrap around to first if at end)
	nextIndex := (currentIndex + 1) % len(weights)
	return weights[nextIndex].Username, weights[nextIndex].Weight
}
