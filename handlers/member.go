package handlers

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"housematee-tgbot/config"
	"housematee-tgbot/models"
	services "housematee-tgbot/services/gsheets"
)

func GetNumberOfMembers(svc *services.GSheets, spreadsheetId string, currentSheetName string) (int, error) {
	// get number of members read range
	numberOfMembersReadRange := currentSheetName + "!" + config.NumberOfMembersCell

	// get number of members data
	numberOfMembersValue, err := svc.GetValue(context.TODO(), spreadsheetId, numberOfMembersReadRange)
	if err != nil {
		logrus.Errorf("failed to get number of members data: %s", err.Error())
		return 0, err
	}

	// convert number of members data to int
	numberOfMembers := cast.ToInt(numberOfMembersValue)
	return numberOfMembers, nil
}

// GetMembers gets the list of members from the spreadsheet
// Columns: O = Username, P = Weight
func GetMembers(svc *services.GSheets, spreadsheetId string, currentSheetName string) ([]models.Member, error) {
	// get number of members
	numberOfMembers, err := GetNumberOfMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return nil, err
	}

	// get members read range (O:P, starting from row 3, skipping header at row 2)
	membersReadRange := fmt.Sprintf("%s!%s%d:%s%d", currentSheetName, config.MembersStartCol, config.MembersStartRow, config.MembersEndCol, config.MembersStartRow+numberOfMembers-1)
	membersResult, err := svc.Get(context.Background(), spreadsheetId, membersReadRange)
	if err != nil {
		logrus.Errorf("failed to get members: %s", err.Error())
		return nil, err
	}

	// convert the result to a slice of members
	// Column O = Username, Column P = Weight
	members := make([]models.Member, 0, numberOfMembers)
	for i, row := range membersResult.Values {
		if len(row) < 2 {
			continue
		}
		username := cast.ToString(row[0])
		weight := cast.ToInt(row[1])
		if weight == 0 {
			weight = 1 // default weight is 1 if not set
		}
		member := models.Member{
			ID:       i + 1,
			Username: username,
			Weight:   weight,
		}
		members = append(members, member)
	}

	logrus.WithFields(logrus.Fields{
		"members": members,
	}).Debug("loaded members with weights")

	return members, nil
}
