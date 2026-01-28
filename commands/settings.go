package commands

import (
	"fmt"
	"sync"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/sirupsen/logrus"
)

// Settings action constants
const (
	SettingsActionPrefix      = "settings."
	SettingsHouseworkReminder = "housework_reminder"
	SettingsReminderToggle    = "reminder_toggle"
	SettingsBack              = "back"
)

// ReminderState manages the runtime state of housework reminders
var (
	reminderEnabled = true
	reminderMutex   sync.RWMutex
)

// IsReminderEnabled returns whether housework reminders are enabled
func IsReminderEnabled() bool {
	reminderMutex.RLock()
	defer reminderMutex.RUnlock()
	return reminderEnabled
}

// SetReminderEnabled sets the reminder state
func SetReminderEnabled(enabled bool) {
	reminderMutex.Lock()
	defer reminderMutex.Unlock()
	reminderEnabled = enabled
}

// ToggleReminder toggles the reminder state and returns the new state
func ToggleReminder() bool {
	reminderMutex.Lock()
	defer reminderMutex.Unlock()
	reminderEnabled = !reminderEnabled
	return reminderEnabled
}

func Settings(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "settings", "command called")
	return showSettingsMainMenu(bot, ctx, false)
}

func showSettingsMainMenu(bot *gotgbot.Bot, ctx *ext.Context, edit bool) error {
	message := "*Settings*\n\nSelect a setting to configure:"

	// Show status indicators in the menu
	reminderStatus := "ON"
	if !IsReminderEnabled() {
		reminderStatus = "OFF"
	}

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: fmt.Sprintf("Housework Reminders [%s]", reminderStatus), CallbackData: "settings.housework_reminder"},
			},
		},
	}

	if edit {
		cb := ctx.Update.CallbackQuery
		_, _, err := cb.Message.EditText(bot, message, &gotgbot.EditMessageTextOpts{
			ParseMode:   "markdown",
			ReplyMarkup: inlineKeyboard,
		})
		return err
	}

	_, err := ctx.EffectiveMessage.Reply(bot, message, &gotgbot.SendMessageOpts{
		ParseMode:   "markdown",
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		return fmt.Errorf("failed to send /settings response: %w", err)
	}
	return nil
}

// HandleSettingsActionCallback handles settings-related callback queries
func HandleSettingsActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	logUserAction(ctx, "settings_callback", fmt.Sprintf("callback: %s", cb.Data))

	var err error
	switch cb.Data {
	case "settings.housework_reminder":
		err = showHouseworkReminderMenu(bot, ctx)
	case "settings.reminder_toggle":
		err = handleReminderToggle(bot, ctx)
	case "settings.back":
		err = showSettingsMainMenu(bot, ctx, true)
	default:
		_, err = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Unknown action",
		})
		return err
	}

	if err != nil {
		return err
	}

	// Acknowledge the callback
	_, err = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{})
	return err
}

func showHouseworkReminderMenu(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	reminderStatus := "ON"
	buttonText := "Turn OFF"
	if !IsReminderEnabled() {
		reminderStatus = "OFF"
		buttonText = "Turn ON"
	}

	message := fmt.Sprintf("*Housework Reminders*\n\nStatus: *%s*\n\nDaily notifications are sent at 18:30 for tasks that are due.", reminderStatus)

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: buttonText, CallbackData: "settings.reminder_toggle"},
			},
			{
				{Text: "<< Back", CallbackData: "settings.back"},
			},
		},
	}

	_, _, err := cb.Message.EditText(bot, message, &gotgbot.EditMessageTextOpts{
		ParseMode:   "markdown",
		ReplyMarkup: inlineKeyboard,
	})
	return err
}

func handleReminderToggle(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	// Toggle the state
	newState := ToggleReminder()

	stateText := "OFF"
	buttonText := "Turn ON"
	if newState {
		stateText = "ON"
		buttonText = "Turn OFF"
	}

	logrus.WithFields(logrus.Fields{
		"user_id":   ctx.EffectiveUser.Id,
		"username":  ctx.EffectiveUser.Username,
		"new_state": stateText,
	}).Info("housework reminders toggled")

	message := fmt.Sprintf("*Housework Reminders*\n\nStatus: *%s*\n\nDaily notifications are sent at 18:30 for tasks that are due.", stateText)

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: buttonText, CallbackData: "settings.reminder_toggle"},
			},
			{
				{Text: "<< Back", CallbackData: "settings.back"},
			},
		},
	}

	_, _, err := cb.Message.EditText(bot, message, &gotgbot.EditMessageTextOpts{
		ParseMode:   "markdown",
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		logrus.Errorf("failed to edit settings message: %s", err.Error())
	}

	return nil
}
