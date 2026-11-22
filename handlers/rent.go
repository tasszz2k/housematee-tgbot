package handlers

import (
	"context"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"google.golang.org/api/sheets/v4"

	"housematee-tgbot/config"
	"housematee-tgbot/enum"
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

	sb.WriteString(fmt.Sprintf("%s *Rent Saved*\n\n", enum.EmojiCheck))
	sb.WriteString(fmt.Sprintf("%s Electric: %s\n", enum.EmojiElectric, utilities.FormatMoney(int(rentData.Electric))))
	sb.WriteString(fmt.Sprintf("%s Water: %s\n", enum.EmojiWater, utilities.FormatMoney(int(rentData.Water))))
	sb.WriteString(fmt.Sprintf("%s Other fees: %s\n", enum.EmojiPage, utilities.FormatMoney(int(rentData.OtherFees))))
	sb.WriteString(fmt.Sprintf("%s *Total rent: %s*\n", enum.EmojiMoney, utilities.FormatMoney(int(rentData.TotalBill))))
	sb.WriteString(fmt.Sprintf("%s Paid by: %s\n", enum.EmojiUser, rentData.Payer))

	if len(rentData.MemberShares) > 0 {
		sb.WriteString(fmt.Sprintf("\n---\n%s *Share per member*\n", enum.EmojiUsers))
		for _, share := range rentData.MemberShares {
			sb.WriteString(fmt.Sprintf("\n%s *%s*\n", enum.EmojiUser, share.Username))
			sb.WriteString(fmt.Sprintf("\u2022 %s Electric: %s\n", enum.EmojiElectric, utilities.FormatMoney(int(share.ElectricShare))))
			sb.WriteString(fmt.Sprintf("\u2022 %s Water: %s\n", enum.EmojiWater, utilities.FormatMoney(int(share.WaterShare))))
			sb.WriteString(fmt.Sprintf("\u2022 %s Other: %s\n", enum.EmojiPage, utilities.FormatMoney(int(share.OtherShare))))
			sb.WriteString(fmt.Sprintf("\u2022 %s *Total: %s*\n", enum.EmojiMoney, utilities.FormatMoney(int(share.TotalShare))))
		}
	}

	return sb.String()
}
