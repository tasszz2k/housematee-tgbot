package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/sheets/v4"

	"housematee-tgbot/config"
	"housematee-tgbot/models"
	"housematee-tgbot/utilities"
)

// SaveRentData writes rent data to Google Sheets and calculates member shares
// Writes to cells: J5 (Electric), J6 (Water), J7 (Other Fees), J8 (Total), M8 (Payer)
func SaveRentData(rentData *models.RentData) error {
	svc, spreadsheetId, currentSheetName, err := GetCurrentSheetInfo()
	if err != nil {
		return err
	}

	// Get members with weights to calculate shares
	members, err := GetMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		logrus.Warnf("failed to get members for share calculation: %s", err.Error())
		// Continue without member shares
	} else {
		// Calculate per-member shares
		rentData.CalculateMemberShares(members)
	}

	// Prepare batch update requests
	updates := []struct {
		cell  string
		value interface{}
	}{
		{config.RentElectricCell, rentData.Electric},
		{config.RentWaterCell, rentData.Water},
		{config.RentOtherFeesCell, rentData.OtherFees},
		{config.RentTotalCell, rentData.TotalBill},
		{config.RentPayerCell, rentData.Payer},
	}

	// Write each value to the sheet
	for _, update := range updates {
		cellRange := fmt.Sprintf("%s!%s", currentSheetName, update.cell)
		_, err = svc.Update(context.TODO(), spreadsheetId, cellRange, &sheets.ValueRange{
			Values: [][]interface{}{{update.value}},
		})
		if err != nil {
			logrus.Errorf("failed to update rent cell %s: %s", update.cell, err.Error())
			return err
		}
	}

	logrus.WithFields(logrus.Fields{
		"total":      rentData.TotalBill,
		"electric":   rentData.Electric,
		"water":      rentData.Water,
		"other_fees": rentData.OtherFees,
		"payer":      rentData.Payer,
	}).Info("rent data saved to Google Sheets")

	return nil
}

// FormatRentSummary formats the rent data for display to user with emojis
func FormatRentSummary(rentData *models.RentData) string {
	var sb strings.Builder

	sb.WriteString("*Rent saved!*\n\n")
	sb.WriteString("*Summary:*\n")
	sb.WriteString("-----------------\n")
	sb.WriteString(fmt.Sprintf("\u26a1 Electric:   %s\n", utilities.FormatMoney(int(rentData.Electric))))
	sb.WriteString(fmt.Sprintf("\U0001F4A7 Water:      %s\n", utilities.FormatMoney(int(rentData.Water))))
	sb.WriteString(fmt.Sprintf("\U0001F4C4 Other Fees: %s\n", utilities.FormatMoney(int(rentData.OtherFees))))
	sb.WriteString("-----------------\n")
	sb.WriteString(fmt.Sprintf("\U0001F4B0 *Total Rent:* %s\n", utilities.FormatMoney(int(rentData.TotalBill))))
	sb.WriteString(fmt.Sprintf("\U0001F464 *Payer:* %s\n", rentData.Payer))

	// Add per-member breakdown if available
	if len(rentData.MemberShares) > 0 {
		sb.WriteString("\n*Per-member breakdown:*\n")
		for _, share := range rentData.MemberShares {
			sb.WriteString(fmt.Sprintf("\n\U0001F464 *%s:*\n", share.Username))
			sb.WriteString(fmt.Sprintf("  \u26a1 Electric: %s\n", utilities.FormatMoney(int(share.ElectricShare))))
			sb.WriteString(fmt.Sprintf("  \U0001F4A7 Water: %s\n", utilities.FormatMoney(int(share.WaterShare))))
			sb.WriteString(fmt.Sprintf("  \U0001F4C4 Other: %s\n", utilities.FormatMoney(int(share.OtherShare))))
			sb.WriteString(fmt.Sprintf("  \U0001F4B0 *Total: %s*\n", utilities.FormatMoney(int(share.TotalShare))))
		}
	}

	return sb.String()
}
