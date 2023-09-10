package commands

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"housematee-tgbot/handlers"
	"housematee-tgbot/models"
	"housematee-tgbot/utilities"
	"log"
	"strings"
)

const (
	HouseworkListCommand   = "housework.list"
	HouseworkAddCommand    = "housework.add"
	HouseworkUpdateCommand = "housework.update"
	HouseworkDeleteCommand = "housework.delete"
)

// Housework handles the /housework command.
func Housework(bot *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("/housework called")
	// show buttons for these commands
	// - Supported commands:
	// - /list - List all housework.
	// - /add - Add new housework.
	// - /update - update a record.
	// - /delete - delete a record.

	// Create an inline keyboard with buttons for each command
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "List", CallbackData: "housework.list"},
				{Text: "Add", CallbackData: "housework.add"},
				{Text: "Update", CallbackData: "housework.update"},
				{Text: "Delete", CallbackData: "housework.delete"},
			},
		},
	}

	// Reply to the user with the available commands as buttons
	_, err := ctx.EffectiveMessage.Reply(bot, "Select a housework action:", &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}
	return nil
}

func HandleHouseworkActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	// Check the CallbackData to determine which button was clicked
	switch cb.Data {
	case HouseworkListCommand:
		// Handle the /splitbill button click
		err := HandleHouseworkListActionCallback(bot, ctx)
		if err != nil {
			return err
		}
	case HouseworkAddCommand:
		// Handle the /housework button click
		err := Todo(bot, ctx)
		if err != nil {
			return err
		}
	case HouseworkUpdateCommand:
		// Handle the /gsheets button click
		err := Todo(bot, ctx)
		if err != nil {
			return err
		}
	case HouseworkDeleteCommand:
		// Handle the /settings button click
		err := Todo(bot, ctx)
		if err != nil {
			return err
		}
	default:
		// Handle other button clicks (if any)
		// Get prefix from CallbackData
		if strings.HasPrefix(cb.Data, HouseworkListCommand) {
			// Handle select the housework to view command
			err := HandleHouseworkSelectActionCallback(bot, ctx)
			if err != nil {
				return err
			}
		}
	}

	// Send a response to acknowledge the button click
	_, err := cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
		Text: fmt.Sprintf("You clicked %s", cb.Data),
	})
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	return nil
}

func HandleHouseworkListActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	// get the list of housework
	houseworkList, err := handlers.GetHouseworkMap()
	if err != nil {
		return err
	}

	// show the list of housework
	// Create an inline keyboard with buttons for each command
	keyboard := make([][]gotgbot.InlineKeyboardButton, 0, len(houseworkList))
	for _, housework := range houseworkList {
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
			{Text: housework.Name, CallbackData: fmt.Sprintf("housework.list.%d", housework.ID)},
		})
	}
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}

	// Reply to the user with the available commands as buttons
	_, err = ctx.EffectiveMessage.Reply(bot, "Select a housework to view:", &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}
	return nil
}

func HandleHouseworkSelectActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	// extract housework id from CallbackData
	// example: housework.list.1, housework.list.2, ...
	houseworkIdStr := strings.Split(ctx.Update.CallbackQuery.Data, ".")[2]
	houseworkId := cast.ToInt(houseworkIdStr)

	// get the list of housework
	houseworkMap, err := handlers.GetHouseworkMap()
	if err != nil {
		return err
	}

	// get the housework
	housework, ok := houseworkMap[houseworkId]
	if !ok {
		return fmt.Errorf("housework with id %d not found", houseworkId)
	}

	// show the housework
	// Create an inline keyboard with buttons for each command
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "Mark as done", CallbackData: fmt.Sprintf("housework.done.%d", housework.ID)},
				{Text: "Assign to housemate", CallbackData: fmt.Sprintf("housework.assign.%d", housework.ID)},
			},
			{
				{Text: "Update", CallbackData: fmt.Sprintf("housework.update.%d", housework.ID)},
				{Text: "Delete", CallbackData: fmt.Sprintf("housework.delete.%d", housework.ID)},
			},
		},
	}

	// Reply to the user with the available commands as buttons
	// Show housework info
	_, err = ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("Housework info:\n---\n%s", handlers.ConvertHouseworkToMarkdownFormat(housework)), &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
		ParseMode:   "markdown",
	})
	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}
	return nil
}

// NotifyDueTasks sends a notification to the channel when there are tasks due today or overdue.
func NotifyDueTasks(bot *gotgbot.Bot) {
	// get all tasks
	houseworkMap, err := handlers.GetHouseworkMap()
	if err != nil {
		logrus.Errorf("failed to get housework map: %s", err.Error())
		return
	}

	// get the list of tasks due today or overdue.
	tasksDueToday := make([]models.Task, 0)
	for _, housework := range houseworkMap {
		isDateDueOrOverdue, _ := utilities.IsDateDueOrOverdue(housework.NextDue)
		if isDateDueOrOverdue {
			tasksDueToday = append(tasksDueToday, housework)
		}
	}

	// if there is no task due today, return
	if len(tasksDueToday) == 0 {
		return
	}

	// send notification to the channel

	for _, task := range tasksDueToday {
		// get the channel id from the task
		channelId := task.ChannelId

		// send the message to the channel
		_, err = bot.SendMessage(channelId, fmt.Sprintf("📢 *%s* is due today!\n*Assignee:* %s\n", task.Name, task.Assignee), &gotgbot.SendMessageOpts{
			ParseMode: "markdown",
			ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
				InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
					{
						{Text: task.Name, CallbackData: fmt.Sprintf("housework.list.%d", task.ID)},
					},
				},
			},
		})
		if err != nil {
			logrus.Errorf("failed to send message to channel %d: %s", channelId, err.Error())
		}
	}
}
