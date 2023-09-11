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
// return map[username]Member
func GetMembers(svc *services.GSheets, spreadsheetId string, currentSheetName string) ([]models.Member, error) {
	// get number of members
	numberOfMembers, err := GetNumberOfMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return nil, err
	}

	// get members read range
	membersReadRange := fmt.Sprintf("%s!%s%d:%s%d", currentSheetName, config.MembersStartCol, config.MembersStartRow, config.MembersEndCol, config.MembersStartRow+numberOfMembers)
	membersResult, err := svc.Get(context.Background(), spreadsheetId, membersReadRange)
	if err != nil {
		logrus.Errorf("failed to get members: %s", err.Error())
		return nil, err
	}

	// map result to the fixed length array
	values := make([][2]string, numberOfMembers)
	for i := 1; i < len(membersResult.Values); i++ {
		for j := 0; j < len(membersResult.Values[i]); j++ {
			values[i-1][j] = cast.ToString(membersResult.Values[i][j])
		}
	}

	// convert the result to a map of members with key is the member id
	members := make([]models.Member, 0, numberOfMembers)
	for _, value := range values {
		member := models.Member{
			ID:       cast.ToInt(value[0]),
			Username: value[1],
		}
		members = append(members, member)
	}
	return members, nil
}
