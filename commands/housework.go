package commands

import (
	"fmt"
	"log"
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
	log.Println("/housework called")
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
		bot, "Select a housework to view:", &gotgbot.SendMessageOpts{
			ReplyMarkup: inlineKeyboard,
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
	svc, spreadsheetId, currentSheetName, err := handlers.GetCurrentSheetInfo()
	if err != nil {
		return err
	}
	// get the member list
	members, err := handlers.GetMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return err
	}

	// assign to the next assignee
	// map[username]member = [
	// 	"username1": member1(id=1),
	//	"username2": member2(id=2),
	//  "username3": members3(id=3)
	//	]
	// case1:
	// currentAssignee = "username1"(id=1)
	// => nextAssignee = "username2"(id=2)
	// case2:
	// currentAssignee = "username3"(id=3) (= last member)
	// => nextAssignee = "username1"(id=1)

	// get the current assignee
	currentAssignee := housework.Assignee
	numOfMembers := len(members)
	var nextAssignee string
	for i, member := range members {
		if member.Username == currentAssignee {
			if i == numOfMembers-1 {
				// last member
				nextAssignee = members[0].Username
			} else {
				nextAssignee = members[i+1].Username
			}
			break
		}
	}

	// update only the assignee
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
	svc, spreadsheetId, currentSheetName, err := handlers.GetCurrentSheetInfo()
	if err != nil {
		return err
	}
	// get the member list
	members, err := handlers.GetMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return err
	}

	// assign to the next assignee
	// map[username]member = [
	// 	"username1": member1(id=1),
	//	"username2": member2(id=2),
	//  "username3": members3(id=3)
	//	]
	// case1:
	// currentAssignee = "username1"(id=1)
	// => nextAssignee = "username2"(id=2)
	// case2:
	// currentAssignee = "username3"(id=3) (= last member)
	// => nextAssignee = "username1"(id=1)

	// get the current assignee
	currentAssignee := housework.Assignee
	numOfMembers := len(members)
	var nextAssignee string
	for i, member := range members {
		if member.Username == currentAssignee {
			if i == numOfMembers-1 {
				// last member
				nextAssignee = members[0].Username
			} else {
				nextAssignee = members[i+1].Username
			}
			break
		}
	}

	// update the housework
	housework.Assignee = nextAssignee

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
		_, err = bot.SendMessage(
			channelId,
			fmt.Sprintf("📢 *%s* is due today!\n*Assignee:* %s\n", task.Name, task.Assignee),
			&gotgbot.SendMessageOpts{
				ParseMode: "markdown",
				ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text:         task.Name,
								CallbackData: fmt.Sprintf("housework.%d.view", task.ID),
							},
						},
					},
				},
			},
		)
		if err != nil {
			logrus.Errorf("failed to send message to channel %d: %s", channelId, err.Error())
		}
	}
}
