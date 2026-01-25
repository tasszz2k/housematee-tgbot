package handlers

import (
	"context"
	"github.com/sirupsen/logrus"
	"housematee-tgbot/config"
	services "housematee-tgbot/services/gsheets"
)

func GetCurrentSheetInfo() (svc *services.GSheets, spreadsheetId string, currentSheetName string, err error) {
	svc = services.GetGSheetsSvc()
	spreadsheetId = config.GetAppConfig().GoogleSheets.SpreadsheetId

	// get current sheet name
	logrus.Infof("Reading current sheet from: %s, cell: %s", spreadsheetId, config.CurrentSheetNameCell)
	currentSheetName, err = svc.GetValue(
		context.TODO(),
		spreadsheetId,
		config.CurrentSheetNameCell,
	)
	if err != nil {
		logrus.Errorf("failed to get current sheet name: %s", err.Error())
		return
	}
	logrus.Infof("Current sheet name value: %s", currentSheetName)
	return
}
