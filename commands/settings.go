package commands

import (
	"fmt"
	"sync"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/sirupsen/logrus"

	"housematee-tgbot/enum"
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
	message := enum.EmojiSettings + " *Settings*\n\nChoose a setting:"

	reminderIcon, reminderStatus := enum.EmojiBell, "ON"
	if !IsReminderEnabled() {
		reminderIcon, reminderStatus = enum.EmojiBellOff, "OFF"
	}

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: fmt.Sprintf("%s Housework Reminders: %s", reminderIcon, reminderStatus), CallbackData: "settings.housework_reminder"},
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
	return renderHouseworkReminderMenu(bot, ctx.Update.CallbackQuery, IsReminderEnabled())
}

func renderHouseworkReminderMenu(bot *gotgbot.Bot, cb *gotgbot.CallbackQuery, enabled bool) error {
	status := enum.EmojiOn + " *ON*"
	button := enum.EmojiBellOff + " Turn OFF"
	if !enabled {
		status = enum.EmojiOff + " *OFF*"
		button = enum.EmojiBell + " Turn ON"
	}

	message := fmt.Sprintf(
		"%s *Housework Reminders*\n\nStatus: %s\n\nWhen ON, the bot sends a reminder at 18:30 every day for each task that is due.",
		enum.EmojiBell, status,
	)

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: button, CallbackData: "settings.reminder_toggle"},
			},
			{
				{Text: enum.EmojiBack + " Back", CallbackData: "settings.back"},
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
	newState := ToggleReminder()

	logrus.WithFields(logrus.Fields{
		"user_id":   ctx.EffectiveUser.Id,
		"username":  ctx.EffectiveUser.Username,
		"new_state": newState,
	}).Info("housework reminders toggled")

	if err := renderHouseworkReminderMenu(bot, ctx.Update.CallbackQuery, newState); err != nil {
		logrus.Errorf("failed to edit settings message: %s", err.Error())
	}
	return nil
}
