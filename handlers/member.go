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
// Columns: O = ID, P = Username, Q = Weight
func GetMembers(svc *services.GSheets, spreadsheetId string, currentSheetName string) ([]models.Member, error) {
	// get number of members
	numberOfMembers, err := GetNumberOfMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return nil, err
	}

	// get members read range (O:Q, starting from row 4, row 3 is header)
	membersReadRange := fmt.Sprintf("%s!%s%d:%s%d", currentSheetName, config.MembersStartCol, config.MembersStartRow, config.MembersEndCol, config.MembersStartRow+numberOfMembers-1)
	logrus.Debugf("reading members from range: %s", membersReadRange)

	membersResult, err := svc.Get(context.Background(), spreadsheetId, membersReadRange)
	if err != nil {
		logrus.Errorf("failed to get members: %s", err.Error())
		return nil, err
	}

	// convert the result to a slice of members
	// Column O = ID, Column P = Username, Column Q = Weight
	members := make([]models.Member, 0, numberOfMembers)
	for _, row := range membersResult.Values {
		if len(row) < 2 {
			continue
		}
		id := cast.ToInt(row[0])
		username := cast.ToString(row[1])
		weight := 1 // default weight
		if len(row) >= 3 {
			weight = cast.ToInt(row[2])
			if weight == 0 {
				weight = 1
			}
		}
		member := models.Member{
			ID:       id,
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
