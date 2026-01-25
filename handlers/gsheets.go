package handlers

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/sheets/v4"

	"housematee-tgbot/config"
	services "housematee-tgbot/services/gsheets"
)

// SheetInfo contains information about a created sheet
type SheetInfo struct {
	SheetName string
	SheetId   int64
}

// GetCurrentSheetName returns the current sheet name from Database!B2
func GetCurrentSheetName() (string, error) {
	svc, spreadsheetId, currentSheetName, err := GetCurrentSheetInfo()
	if err != nil {
		return "", err
	}
	_ = svc // unused but needed for GetCurrentSheetInfo call
	_ = spreadsheetId
	return currentSheetName, nil
}

// GetSheetIdByName finds a sheet's numeric ID by its name
func GetSheetIdByName(svc *services.GSheets, spreadsheetId string, sheetName string) (int64, error) {
	spreadsheet, err := svc.GetSpreadsheet(context.TODO(), spreadsheetId)
	if err != nil {
		logrus.Errorf("failed to get spreadsheet: %s", err.Error())
		return 0, err
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return sheet.Properties.SheetId, nil
		}
	}

	return 0, fmt.Errorf("sheet '%s' not found", sheetName)
}

// SheetExists checks if a sheet with the given name already exists
func SheetExists(svc *services.GSheets, spreadsheetId string, sheetName string) (bool, error) {
	spreadsheet, err := svc.GetSpreadsheet(context.TODO(), spreadsheetId)
	if err != nil {
		logrus.Errorf("failed to get spreadsheet: %s", err.Error())
		return false, err
	}

	for _, sheet := range spreadsheet.Sheets {
		if sheet.Properties.Title == sheetName {
			return true, nil
		}
	}

	return false, nil
}

// CreateNewMonthSheet creates a new sheet by copying the Template and updates Database!B2
func CreateNewMonthSheet(newSheetName string, displayName string) (*SheetInfo, error) {
	svc, spreadsheetId, _, err := GetCurrentSheetInfo()
	if err != nil {
		return nil, err
	}

	// Check if sheet already exists
	exists, err := SheetExists(svc, spreadsheetId, newSheetName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("sheet '%s' already exists", newSheetName)
	}

	// Get the Template sheet ID
	templateSheetId, err := GetSheetIdByName(svc, spreadsheetId, config.TemplateSheetName)
	if err != nil {
		return nil, fmt.Errorf("failed to find Template sheet: %w", err)
	}

	// Duplicate the Template sheet with the new name
	newSheetProps, err := svc.DuplicateSheet(context.TODO(), spreadsheetId, templateSheetId, newSheetName)
	if err != nil {
		logrus.Errorf("failed to duplicate sheet: %s", err.Error())
		return nil, err
	}

	// Update cell A1 in the new sheet with the display name (MM/YYYY)
	err = updateSheetCell(svc, spreadsheetId, newSheetName, "A1", displayName)
	if err != nil {
		logrus.Errorf("failed to update A1 cell: %s", err.Error())
		return nil, err
	}

	// Update Database!B2 with the new sheet name
	err = updateCurrentSheetName(svc, spreadsheetId, newSheetName)
	if err != nil {
		logrus.Errorf("failed to update current sheet name: %s", err.Error())
		return nil, err
	}

	return &SheetInfo{
		SheetName: newSheetName,
		SheetId:   newSheetProps.SheetId,
	}, nil
}

// updateCurrentSheetName updates the Database!B2 cell with the new sheet name
func updateCurrentSheetName(svc *services.GSheets, spreadsheetId string, sheetName string) error {
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{{sheetName}},
	}

	_, err := svc.Update(context.TODO(), spreadsheetId, config.CurrentSheetNameCell, valueRange)
	if err != nil {
		return err
	}

	return nil
}

// updateSheetCell updates a specific cell in a sheet
func updateSheetCell(svc *services.GSheets, spreadsheetId string, sheetName string, cell string, value string) error {
	cellRange := fmt.Sprintf("%s!%s", sheetName, cell)
	valueRange := &sheets.ValueRange{
		Values: [][]interface{}{{value}},
	}

	_, err := svc.Update(context.TODO(), spreadsheetId, cellRange, valueRange)
	if err != nil {
		return err
	}

	return nil
}
