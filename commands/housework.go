package commands

import (
	"fmt"
	"strings"

	"housematee-tgbot/enum"
	"housematee-tgbot/handlers"
	"housematee-tgbot/models"
	"housematee-tgbot/utilities"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
)

const (
	HouseworkListCommand   = "housework.list"
	HouseworkAddCommand    = "housework.add"
	HouseworkUpdateCommand = "housework.update"
	HouseworkDeleteCommand = "housework.delete"

	HouseworkActionPrefix   = "housework."
	HouseworkViewAction     = "view"
	HouseworkMarkDoneAction = "done"
	HouseworkAssignAction   = "assign"
	HouseworkUpdateAction   = "update"
	HouseworkDeleteAction   = "delete"
)

// Housework handles the /housework command.
func Housework(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "housework", "command called")
	// show buttons for these commands
	// - Supported commands:
	// - /list - List all housework.
	// - /add - Add new housework.
	// - /update - update a record.
	// - /delete - delete a record.

	//// Create an inline keyboard with buttons for each command
	//inlineKeyboard := gotgbot.InlineKeyboardMarkup{
	//	InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
	//		{
	//			{Text: "List", CallbackData: "housework.list"},
	//			{Text: "Add", CallbackData: "housework.add"},
	//			//{Text: "Update", CallbackData: "housework.update"},
	//			//{Text: "Delete", CallbackData: "housework.delete"},
	//		},
	//	},
	//}
	//
	//// Reply to the user with the available commands as buttons
	//_, err := ctx.EffectiveMessage.Reply(bot, "Select a housework action:", &gotgbot.SendMessageOpts{
	//	ReplyMarkup: inlineKeyboard,
	//})

	// show list of housework
	err := HandleHouseworkListActionCallback(bot, ctx)

	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}
	return nil
}

func HandleHouseworkActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	logUserAction(ctx, "housework_callback", fmt.Sprintf("callback: %s", cb.Data))

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
		if strings.HasPrefix(cb.Data, HouseworkActionPrefix) {
			// Handle select the housework to view command
			err := HandleHouseworkSelectActionCallback(bot, ctx)
			if err != nil {
				return err
			}
		}
	}

	// Send a response to acknowledge the button click
	_, err := cb.Answer(
		bot, &gotgbot.AnswerCallbackQueryOpts{
			Text: fmt.Sprintf("You clicked %s", cb.Data),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	return nil
}

func HandleHouseworkListActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "housework_list", "listing housework tasks")
	// get the list of housework
	houseworkList, err := handlers.GetHouseworkMap()
	if err != nil {
		return err
	}

	// show the list of housework
	// Create an inline keyboard with buttons for each command
	keyboard := make([][]gotgbot.InlineKeyboardButton, 0, len(houseworkList))
	for _, housework := range houseworkList {
		name := housework.Name
		isDateDueOrOverdue, _ := utilities.IsDateDueOrOverdue(housework.NextDue)
		if isDateDueOrOverdue {
			name += " » 📢"
		}
		keyboard = append(
			keyboard, []gotgbot.InlineKeyboardButton{
				{Text: name, CallbackData: fmt.Sprintf("housework.%d.view", housework.ID)},
			},
		)
	}
	keyboard = append(
		keyboard, []gotgbot.InlineKeyboardButton{
			{Text: "➕ Add new housework", CallbackData: "housework.add"},
		},
	)

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}

	// Reply to the user with the available commands as buttons
	_, err = ctx.EffectiveMessage.Reply(
		bot, "*Housework Tasks*\n\nSelect a task to view details:", &gotgbot.SendMessageOpts{
			ReplyMarkup: inlineKeyboard,
			ParseMode:   "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}
	return nil
}

func HandleHouseworkSelectActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	// extract housework id from CallbackData
	// [object].[id].[action]
	// example: housework.1.view, housework.2.view, ...
	commandElements := strings.Split(ctx.Update.CallbackQuery.Data, ".")
	houseworkIdStr := commandElements[1]
	houseworkId := cast.ToInt(houseworkIdStr)
	selectedAction := commandElements[2]

	logUserAction(ctx, "housework_select", fmt.Sprintf("task_id=%d action=%s", houseworkId, selectedAction))

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

	numberOfHousework := len(houseworkMap)

	switch selectedAction {
	case HouseworkViewAction:
		// show the housework
		err = handleHouseworkViewAction(bot, ctx, housework, "Housework info")
	case HouseworkMarkDoneAction:
		// mark the housework as done
		err = handleHouseworkMarkDoneAction(bot, ctx, housework, numberOfHousework)
	case HouseworkAssignAction:
		// assign the housework to other
		err = handleHouseworkAssignToOtherAction(bot, ctx, housework, numberOfHousework)
	}

	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}

	return nil
}

func MarkAsDoneHouseworkByShortcut(bot *gotgbot.Bot, ctx *ext.Context) error {
	// extract housework id from command
	// command format: "hw[housework_id]"
	command := getCommandFromMessage(bot, ctx.Message)
	if !strings.HasPrefix(command, enum.HouseworkPrefix) {
		return fmt.Errorf("invalid the marking as done housework command: %s", command)
	}

	houseworkIdStr := strings.TrimPrefix(command, enum.HouseworkPrefix)
	houseworkId := cast.ToInt(houseworkIdStr)

	logUserAction(ctx, "housework_shortcut", fmt.Sprintf("mark_done task_id=%d", houseworkId))

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

	numberOfHousework := len(houseworkMap)
	err = handleHouseworkMarkDoneAction(bot, ctx, housework, numberOfHousework)
	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}

	return nil
}

func handleHouseworkAssignToOtherAction(
	bot *gotgbot.Bot,
	ctx *ext.Context,
	housework models.Task,
	numberOfHousework int,
) error {
	logUserAction(ctx, "housework_assign", fmt.Sprintf("task_id=%d task_name=%s current_assignee=%s", housework.ID, housework.Name, housework.Assignee))

	svc, spreadsheetId, currentSheetName, err := handlers.GetCurrentSheetInfo()
	if err != nil {
		return err
	}

	// Round-robin rotation using Members list
	members, err := handlers.GetMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return err
	}

	var nextAssignee string
	currentAssignee := housework.Assignee
	numOfMembers := len(members)
	for i, member := range members {
		if member.Username == currentAssignee {
			if i == numOfMembers-1 {
				nextAssignee = members[0].Username
			} else {
				nextAssignee = members[i+1].Username
			}
			break
		}
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       ctx.EffectiveUser.Id,
		"task_id":       housework.ID,
		"prev_assignee": housework.Assignee,
		"next_assignee": nextAssignee,
	}).Info("assigned to other member")

	housework.Assignee = nextAssignee

	// upsert the housework
	err = handlers.UpdateHousework(
		svc,
		spreadsheetId,
		currentSheetName,
		housework,
		numberOfHousework,
	)
	if err != nil {
		return err
	}

	// show the housework
	err = handleHouseworkViewAction(bot, ctx, housework, "Housework is assigned to other")
	if err != nil {
		return err
	}

	return nil
}

func handleHouseworkMarkDoneAction(
	bot *gotgbot.Bot,
	ctx *ext.Context,
	housework models.Task,
	numberOfHousework int,
) error {
	logUserAction(ctx, "housework_mark_done", fmt.Sprintf("task_id=%d task_name=%s assignee=%s", housework.ID, housework.Name, housework.Assignee))

	svc, spreadsheetId, currentSheetName, err := handlers.GetCurrentSheetInfo()
	if err != nil {
		return err
	}

	// Round-robin rotation using Members list
	members, err := handlers.GetMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return err
	}

	var nextAssignee string
	currentAssignee := housework.Assignee
	numOfMembers := len(members)
	for i, member := range members {
		if member.Username == currentAssignee {
			if i == numOfMembers-1 {
				nextAssignee = members[0].Username
			} else {
				nextAssignee = members[i+1].Username
			}
			break
		}
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       ctx.EffectiveUser.Id,
		"task_id":       housework.ID,
		"prev_assignee": housework.Assignee,
		"next_assignee": nextAssignee,
	}).Info("rotated to next assignee")

	housework.Assignee = nextAssignee

	// Update LastDone and NextDue
	housework.LastDone = utilities.GetCurrentDate()
	nextDue, err := utilities.AddDay(housework.LastDone, housework.Frequency)
	if err != nil {
		logrus.Errorf("failed to add day: %s", err.Error())
		return err
	}
	housework.NextDue = nextDue

	// upsert the housework
	err = handlers.UpdateHousework(
		svc,
		spreadsheetId,
		currentSheetName,
		housework,
		numberOfHousework,
	)
	if err != nil {
		return err
	}

	// show the housework
	err = handleHouseworkViewAction(bot, ctx, housework, "Housework is updated")
	if err != nil {
		return err
	}

	return nil
}

func handleHouseworkViewAction(
	bot *gotgbot.Bot,
	ctx *ext.Context,
	housework models.Task,
	title string,
) error {
	// Creates an inline keyboard with buttons for each command
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "Mark as done",
					CallbackData: fmt.Sprintf("housework.%d.done", housework.ID),
				},
				{
					Text:         "Assign to other",
					CallbackData: fmt.Sprintf("housework.%d.assign", housework.ID),
				},
			},
			{
				{Text: "Update", CallbackData: fmt.Sprintf("housework.%d.update", housework.ID)},
				{Text: "Delete", CallbackData: fmt.Sprintf("housework.%d.delete", housework.ID)},
			},
		},
	}

	// Reply to the user with the available commands as buttons
	// Show housework info
	_, err := ctx.EffectiveMessage.Reply(
		bot,
		fmt.Sprintf(
			"%s:\n---\n%s",
			title,
			handlers.ConvertHouseworkToMarkdownFormat(housework),
		),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: inlineKeyboard,
			ParseMode:   "markdown",
		},
	)
	return err
}

// NotifyDueTasks sends a notification to the channel when there are tasks due today or overdue.
func NotifyDueTasks(bot *gotgbot.Bot) {
	// Check if reminders are enabled
	if !IsReminderEnabled() {
		logrus.Debug("housework reminders are disabled, skipping notification")
		return
	}

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

		// Build notification message with details
		message := fmt.Sprintf(
			"*%s* is due!\n\n"+
				"*Assignee:* %s\n"+
				"*Due date:* %s",
			task.Name,
			task.Assignee,
			task.NextDue,
		)

		// send the message to the channel
		_, err = bot.SendMessage(
			channelId,
			message,
			&gotgbot.SendMessageOpts{
				ParseMode: "markdown",
				ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text:         "Mark as done",
								CallbackData: fmt.Sprintf("housework.%d.done", task.ID),
							},
							{
								Text:         "View details",
								CallbackData: fmt.Sprintf("housework.%d.view", task.ID),
							},
						},
					},
				},
			},
		)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"channel_id": channelId,
				"task_id":    task.ID,
				"task_name":  task.Name,
			}).Errorf("failed to send notification: %s", err.Error())
		} else {
			logrus.WithFields(logrus.Fields{
				"channel_id": channelId,
				"task_id":    task.ID,
				"task_name":  task.Name,
				"assignee":   task.Assignee,
			}).Info("sent due task notification")
		}
	}
}
