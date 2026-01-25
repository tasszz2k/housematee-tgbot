package commands

import (
	"fmt"
	"html"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"housematee-tgbot/enum"
	"housematee-tgbot/handlers"
	"housematee-tgbot/utilities"
)

// escapeHTML escapes special HTML characters in a string
func escapeHTML(s string) string {
	return html.EscapeString(s)
}

const (
	GSheetsCreateCommand        = "gsheets.create"
	GSheetsConfirmCreateCommand = "gsheets.confirm_create"
	GSheetsCancelCreateCommand  = "gsheets.cancel_create"
)

// GSheets handles the /gsheets command
func GSheets(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "gsheets", "command called")

	// Get current sheet name from Database
	currentSheet, err := handlers.GetCurrentSheetName()
	if err != nil {
		currentSheet = "Unable to fetch"
	}

	// Create an inline keyboard with buttons for each action
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Create New Sheet", CallbackData: GSheetsCreateCommand},
			},
		},
	}

	// Reply to the user with the current sheet info and available actions
	_, err = ctx.EffectiveMessage.Reply(
		bot,
		fmt.Sprintf(
			"*Google Sheets Management*\n\n"+
				"*Current Sheet:* `%s`\n\n"+
				"Select an action:",
			currentSheet,
		),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: inlineKeyboard,
			ParseMode:   "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send /gsheets response: %w", err)
	}
	return nil
}

// HandleGSheetsActionCallback handles all gsheets.* callback queries
func HandleGSheetsActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	logUserAction(ctx, "gsheets_callback", fmt.Sprintf("callback: %s", cb.Data))

	// Check the CallbackData to determine which action was selected
	switch cb.Data {
	case GSheetsCreateCommand:
		err := handleGSheetsCreateAction(bot, ctx)
		if err != nil {
			return err
		}
	case GSheetsConfirmCreateCommand:
		err := handleGSheetsConfirmCreateAction(bot, ctx)
		if err != nil {
			return err
		}
	case GSheetsCancelCreateCommand:
		err := handleGSheetsCancelCreateAction(bot, ctx)
		if err != nil {
			return err
		}
	default:
		// Handle other actions with gsheets. prefix
		if strings.HasPrefix(cb.Data, enum.GSheetsActionPrefix) {
			// Future actions can be handled here
			err := Todo(bot, ctx)
			if err != nil {
				return err
			}
		}
	}

	// Send a response to acknowledge the button click
	_, err := cb.Answer(
		bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: fmt.Sprintf("Processing %s", cb.Data),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	return nil
}

// handleGSheetsCreateAction shows confirmation dialog for creating a new sheet
func handleGSheetsCreateAction(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Get the current month sheet name (YYYY_MM format)
	newSheetName := utilities.GetCurrentMonthSheetName()

	// Create confirmation buttons
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Confirm", CallbackData: GSheetsConfirmCreateCommand},
				{Text: "Cancel", CallbackData: GSheetsCancelCreateCommand},
			},
		},
	}

	// Show confirmation message with the draft sheet name
	_, err := ctx.EffectiveMessage.Reply(
		bot,
		fmt.Sprintf(
			"*Create New Sheet*\n\n"+
				"A new sheet will be created by copying the Template.\n\n"+
				"*Sheet Name:* `%s`\n\n"+
				"Do you want to proceed?",
			newSheetName,
		),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: inlineKeyboard,
			ParseMode:   "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send create confirmation: %w", err)
	}
	return nil
}

// handleGSheetsConfirmCreateAction executes the sheet creation
func handleGSheetsConfirmCreateAction(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Get the current month sheet name (YYYY_MM for sheet name)
	newSheetName := utilities.GetCurrentMonthSheetName()
	// Get the display name (MM/YYYY for cell A1)
	displayName := utilities.GetCurrentMonthDisplayName()

	// Create the new sheet
	sheetInfo, err := handlers.CreateNewMonthSheet(newSheetName, displayName)
	if err != nil {
		// Send error message to user (use HTML to avoid markdown parsing issues with special chars)
		_, sendErr := ctx.EffectiveMessage.Reply(
			bot,
			fmt.Sprintf("<b>Status:</b> Failed\n\n<b>Error:</b> %s", escapeHTML(err.Error())),
			&gotgbot.SendMessageOpts{
				ParseMode: "HTML",
			},
		)
		if sendErr != nil {
			return fmt.Errorf("failed to send error message: %w", sendErr)
		}
		return nil
	}

	// Send success message
	_, err = ctx.EffectiveMessage.Reply(
		bot,
		fmt.Sprintf(
			"*Status:* Success\n\n"+
				"---\n"+
				"*Sheet Name:* `%s`\n"+
				"*Sheet ID:* `%d`",
			sheetInfo.SheetName,
			sheetInfo.SheetId,
		),
		&gotgbot.SendMessageOpts{
			ParseMode: "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send success message: %w", err)
	}
	return nil
}

// handleGSheetsCancelCreateAction cancels the sheet creation
func handleGSheetsCancelCreateAction(bot *gotgbot.Bot, ctx *ext.Context) error {
	_, err := ctx.EffectiveMessage.Reply(
		bot,
		"Sheet creation cancelled.",
		&gotgbot.SendMessageOpts{
			ParseMode: "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send cancel message: %w", err)
	}
	return nil
}
