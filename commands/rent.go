package commands

import (
	"fmt"
	"sync"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	tgBotHandler "github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/spf13/cast"

	"housematee-tgbot/enum"
	"housematee-tgbot/handlers"
	"housematee-tgbot/models"
	"housematee-tgbot/utilities"
)

// rentDataStore stores rent data during conversation (keyed by chat ID)
var (
	rentDataStore = make(map[int64]*models.RentData)
	rentDataMux   sync.RWMutex
)

// getRentData gets rent data for a chat
func getRentData(chatID int64) *models.RentData {
	rentDataMux.RLock()
	defer rentDataMux.RUnlock()
	return rentDataStore[chatID]
}

// setRentData sets rent data for a chat
func setRentData(chatID int64, data *models.RentData) {
	rentDataMux.Lock()
	defer rentDataMux.Unlock()
	rentDataStore[chatID] = data
}

// clearRentData clears rent data for a chat
func clearRentData(chatID int64) {
	rentDataMux.Lock()
	defer rentDataMux.Unlock()
	delete(rentDataStore, chatID)
}

// Rent handles the /rent command - entry point
func Rent(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "rent", "command called")
	return StartRentConversation(bot, ctx)
}

// StartRentConversation starts the rent conversation flow
func StartRentConversation(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "rent_start", "starting rent flow")

	// Initialize rent data for this chat
	setRentData(ctx.EffectiveChat.Id, &models.RentData{})

	_, err := ctx.EffectiveMessage.Reply(
		bot,
		"\U0001F3E0 *Add Rent*\n\nPlease enter the \U0001F4B0 *total rent bill* amount:",
		&gotgbot.SendMessageOpts{ParseMode: "markdown"},
	)
	if err != nil {
		return err
	}
	return tgBotHandler.NextConversationState(enum.RentStateTotal)
}

// HandleRentTotalInput handles the total bill input
func HandleRentTotalInput(bot *gotgbot.Bot, ctx *ext.Context) error {
	if !CheckPermission(bot, ctx) {
		return nil
	}

	input := ctx.EffectiveMessage.Text
	amount := utilities.ParseAmount(input)

	if amount == "" || !utilities.IsNumeric(amount) {
		_, err := ctx.EffectiveMessage.Reply(
			bot,
			"*Invalid Amount*\n\nPlease enter a valid number for the total bill:",
			&gotgbot.SendMessageOpts{ParseMode: "markdown"},
		)
		if err != nil {
			return err
		}
		return tgBotHandler.NextConversationState(enum.RentStateTotal)
	}

	// Store total bill
	rentData := getRentData(ctx.EffectiveChat.Id)
	if rentData == nil {
		rentData = &models.RentData{}
	}
	rentData.TotalBill = cast.ToInt64(amount)
	setRentData(ctx.EffectiveChat.Id, rentData)

	_, err := ctx.EffectiveMessage.Reply(
		bot,
		fmt.Sprintf("\U0001F4B0 Total: *%s*\n\nNow enter the \u26a1 *electric* bill amount:", utilities.FormatMoney(int(rentData.TotalBill))),
		&gotgbot.SendMessageOpts{ParseMode: "markdown"},
	)
	if err != nil {
		return err
	}
	return tgBotHandler.NextConversationState(enum.RentStateElectric)
}

// HandleRentElectricInput handles the electric bill input
func HandleRentElectricInput(bot *gotgbot.Bot, ctx *ext.Context) error {
	if !CheckPermission(bot, ctx) {
		return nil
	}

	input := ctx.EffectiveMessage.Text
	amount := utilities.ParseAmount(input)

	if amount == "" || !utilities.IsNumeric(amount) {
		_, err := ctx.EffectiveMessage.Reply(
			bot,
			"*Invalid Amount*\n\nPlease enter a valid number for the electric bill:",
			&gotgbot.SendMessageOpts{ParseMode: "markdown"},
		)
		if err != nil {
			return err
		}
		return tgBotHandler.NextConversationState(enum.RentStateElectric)
	}

	// Store electric bill
	rentData := getRentData(ctx.EffectiveChat.Id)
	if rentData == nil {
		clearRentData(ctx.EffectiveChat.Id)
		return StartRentConversation(bot, ctx)
	}
	rentData.Electric = cast.ToInt64(amount)
	setRentData(ctx.EffectiveChat.Id, rentData)

	_, err := ctx.EffectiveMessage.Reply(
		bot,
		fmt.Sprintf("\u26a1 Electric: *%s*\n\nNow enter the \U0001F4A7 *water* bill amount:", utilities.FormatMoney(int(rentData.Electric))),
		&gotgbot.SendMessageOpts{ParseMode: "markdown"},
	)
	if err != nil {
		return err
	}
	return tgBotHandler.NextConversationState(enum.RentStateWater)
}

// HandleRentWaterInput handles the water bill input and saves the rent data
func HandleRentWaterInput(bot *gotgbot.Bot, ctx *ext.Context) error {
	if !CheckPermission(bot, ctx) {
		return nil
	}

	input := ctx.EffectiveMessage.Text
	amount := utilities.ParseAmount(input)

	if amount == "" || !utilities.IsNumeric(amount) {
		_, err := ctx.EffectiveMessage.Reply(
			bot,
			"*Invalid Amount*\n\nPlease enter a valid number for the water bill:",
			&gotgbot.SendMessageOpts{ParseMode: "markdown"},
		)
		if err != nil {
			return err
		}
		return tgBotHandler.NextConversationState(enum.RentStateWater)
	}

	// Store water bill
	rentData := getRentData(ctx.EffectiveChat.Id)
	if rentData == nil {
		clearRentData(ctx.EffectiveChat.Id)
		return StartRentConversation(bot, ctx)
	}
	rentData.Water = cast.ToInt64(amount)

	// Auto-fill payer with current user (who sent the command)
	rentData.Payer = "@" + ctx.EffectiveUser.Username

	// Calculate other fees
	rentData.CalculateOtherFees()

	// Validate: other fees should not be negative
	if rentData.OtherFees < 0 {
		_, err := ctx.EffectiveMessage.Reply(
			bot,
			fmt.Sprintf("*Validation Error*\n\nElectric (%s) + Water (%s) exceeds Total (%s).\n\nPlease start over with /rent",
				utilities.FormatMoney(int(rentData.Electric)),
				utilities.FormatMoney(int(rentData.Water)),
				utilities.FormatMoney(int(rentData.TotalBill)),
			),
			&gotgbot.SendMessageOpts{ParseMode: "markdown"},
		)
		clearRentData(ctx.EffectiveChat.Id)
		if err != nil {
			return err
		}
		return tgBotHandler.EndConversation()
	}

	// Save to Google Sheets
	err := handlers.SaveRentData(rentData)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(
			bot,
			fmt.Sprintf("*Failed to Save Rent*\n\n%s", err.Error()),
			&gotgbot.SendMessageOpts{ParseMode: "markdown"},
		)
		clearRentData(ctx.EffectiveChat.Id)
		if err != nil {
			return err
		}
		return tgBotHandler.EndConversation()
	}

	// Send success message with summary
	summary := handlers.FormatRentSummary(rentData)
	_, err = ctx.EffectiveMessage.Reply(
		bot,
		summary,
		&gotgbot.SendMessageOpts{ParseMode: "markdown"},
	)

	// Clear rent data
	clearRentData(ctx.EffectiveChat.Id)

	if err != nil {
		return err
	}
	return tgBotHandler.EndConversation()
}

// HandleRentPayerInput is deprecated - payer is now auto-filled in HandleRentWaterInput
