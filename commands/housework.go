package commands

import (
	"fmt"
	"sort"
	"strings"
	"sync"

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

// The dispatcher handles updates concurrently, so without houseworkMu two fast
// clicks can both read the old row and both rotate the task.
var (
	houseworkMu              sync.Mutex
	handledHouseworkMessages = make(map[string]struct{})
)

// Only recent messages matter for double-click protection, so the set is reset
// instead of growing forever.
const maxHandledHouseworkMessages = 500

type callbackAnswer struct {
	Text  string
	Alert bool
}

type houseworkAudit struct {
	Action       string
	By           string
	At           string
	FromAssignee string
	ToAssignee   string
}

// Housework handles the /housework command.
func Housework(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "housework", "command called")
	if err := HandleHouseworkListActionCallback(bot, ctx); err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}
	return nil
}

func HandleHouseworkActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	logUserAction(ctx, "housework_callback", fmt.Sprintf("callback: %s", cb.Data))

	var answer callbackAnswer
	var err error
	switch cb.Data {
	case HouseworkListCommand:
		err = HandleHouseworkListActionCallback(bot, ctx)
	case HouseworkAddCommand, HouseworkUpdateCommand, HouseworkDeleteCommand:
		err = Todo(bot, ctx)
	default:
		if strings.HasPrefix(cb.Data, HouseworkActionPrefix) {
			answer, err = handleHouseworkSelectAction(bot, ctx)
		}
	}

	if err != nil {
		_, _ = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
			Text:      "Something went wrong. Please try again.",
			ShowAlert: true,
		})
		return err
	}

	// Telegram accepts only one answer per callback, so sub-handlers return the
	// answer instead of sending it.
	_, err = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{
		Text:      answer.Text,
		ShowAlert: answer.Alert,
	})
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}
	return nil
}

func HandleHouseworkListActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "housework_list", "listing housework tasks")
	houseworkMap, err := handlers.GetHouseworkMap()
	if err != nil {
		return err
	}
	tasks := sortedTasks(houseworkMap)

	keyboard := make([][]gotgbot.InlineKeyboardButton, 0, len(tasks)+1)
	for _, task := range tasks {
		keyboard = append(
			keyboard, []gotgbot.InlineKeyboardButton{
				{Text: houseworkListButtonLabel(task), CallbackData: fmt.Sprintf("housework.%d.view", task.ID)},
			},
		)
	}
	keyboard = append(
		keyboard, []gotgbot.InlineKeyboardButton{
			{Text: enum.EmojiAdd + " Add new housework", CallbackData: HouseworkAddCommand},
		},
	)

	text := enum.EmojiBroom + " *Housework Tasks*\n\n"
	if len(tasks) == 0 {
		text += "No tasks yet."
	} else {
		text += fmt.Sprintf("Tap a task to see the details.\n%s = due today or overdue", enum.EmojiBell)
	}

	_, err = ctx.EffectiveMessage.Reply(
		bot, text, &gotgbot.SendMessageOpts{
			ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyboard},
			ParseMode:   "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}
	return nil
}

func sortedTasks(houseworkMap map[int]models.Task) []models.Task {
	tasks := make([]models.Task, 0, len(houseworkMap))
	for _, task := range houseworkMap {
		tasks = append(tasks, task)
	}
	sort.Slice(tasks, func(i, j int) bool { return tasks[i].ID < tasks[j].ID })
	return tasks
}

// houseworkListButtonLabel builds "icon Name » @assignee » due".
func houseworkListButtonLabel(task models.Task) string {
	icon := enum.EmojiBroom
	due := task.NextDue
	if days, err := utilities.DaysUntil(task.NextDue); err == nil {
		switch {
		case days < 0:
			icon = enum.EmojiBell
			due = fmt.Sprintf("overdue %dd", -days)
		case days == 0:
			icon = enum.EmojiBell
			due = "today"
		default:
			due = task.NextDue[:5]
		}
	}

	parts := []string{icon + " " + task.Name}
	if task.Assignee != "" {
		parts = append(parts, task.Assignee)
	}
	if due != "" {
		parts = append(parts, due)
	}
	return strings.Join(parts, " » ")
}

func handleHouseworkSelectAction(bot *gotgbot.Bot, ctx *ext.Context) (callbackAnswer, error) {
	cb := ctx.Update.CallbackQuery
	// [object].[id].[action], for example housework.1.view
	commandElements := strings.Split(cb.Data, ".")
	if len(commandElements) < 3 {
		return callbackAnswer{}, fmt.Errorf("invalid housework callback data: %s", cb.Data)
	}
	houseworkId := cast.ToInt(commandElements[1])
	selectedAction := commandElements[2]

	logUserAction(ctx, "housework_select", fmt.Sprintf("task_id=%d action=%s", houseworkId, selectedAction))

	switch selectedAction {
	case HouseworkViewAction:
		housework, _, err := getHousework(houseworkId)
		if err != nil {
			return callbackAnswer{}, err
		}
		return callbackAnswer{}, sendHouseworkView(bot, ctx, housework, enum.EmojiDetails, "Housework Details", nil)
	case HouseworkMarkDoneAction, HouseworkAssignAction:
	default:
		// Old messages can still carry the removed Update and Delete buttons.
		return callbackAnswer{}, Todo(bot, ctx)
	}

	houseworkMu.Lock()
	defer houseworkMu.Unlock()

	msgKey := callbackMessageKey(cb)
	if isHouseworkMessageHandled(msgKey) {
		return callbackAnswer{
			Text:  "This message is already handled. Use the latest message.",
			Alert: true,
		}, nil
	}

	housework, numberOfHousework, err := getHousework(houseworkId)
	if err != nil {
		return callbackAnswer{}, err
	}

	var answer callbackAnswer
	var applied bool
	if selectedAction == HouseworkMarkDoneAction {
		answer, applied, err = handleHouseworkMarkDoneAction(bot, ctx, housework, numberOfHousework)
	} else {
		answer, applied, err = handleHouseworkAssignToOtherAction(bot, ctx, housework, numberOfHousework)
	}

	// The sheet is already written when applied is true, so lock the message
	// even if sending the reply failed.
	if applied {
		markHouseworkMessageHandled(msgKey)
		removeMessageButtons(bot, cb)
	}
	return answer, err
}

func MarkAsDoneHouseworkByShortcut(bot *gotgbot.Bot, ctx *ext.Context) error {
	// command format: "hw[housework_id]"
	command := getCommandFromMessage(bot, ctx.Message)
	if !strings.HasPrefix(command, enum.HouseworkPrefix) {
		return fmt.Errorf("invalid the marking as done housework command: %s", command)
	}
	houseworkId := cast.ToInt(strings.TrimPrefix(command, enum.HouseworkPrefix))

	logUserAction(ctx, "housework_shortcut", fmt.Sprintf("mark_done task_id=%d", houseworkId))

	houseworkMu.Lock()
	defer houseworkMu.Unlock()

	housework, numberOfHousework, err := getHousework(houseworkId)
	if err != nil {
		return err
	}

	answer, applied, err := handleHouseworkMarkDoneAction(bot, ctx, housework, numberOfHousework)
	if err != nil {
		return fmt.Errorf("failed to send /housework response: %w", err)
	}
	if !applied {
		_, err = ctx.EffectiveMessage.Reply(bot, enum.EmojiLock+" "+answer.Text, nil)
		return err
	}
	return nil
}

func getHousework(houseworkId int) (models.Task, int, error) {
	houseworkMap, err := handlers.GetHouseworkMap()
	if err != nil {
		return models.Task{}, 0, err
	}
	housework, ok := houseworkMap[houseworkId]
	if !ok {
		return models.Task{}, 0, fmt.Errorf("housework with id %d not found", houseworkId)
	}
	return housework, len(houseworkMap), nil
}

func callbackMessageKey(cb *gotgbot.CallbackQuery) string {
	if cb == nil || cb.Message == nil {
		return ""
	}
	return fmt.Sprintf("%d:%d", cb.Message.GetChat().Id, cb.Message.GetMessageId())
}

func isHouseworkMessageHandled(key string) bool {
	if key == "" {
		return false
	}
	_, ok := handledHouseworkMessages[key]
	return ok
}

func markHouseworkMessageHandled(key string) {
	if key == "" {
		return
	}
	if len(handledHouseworkMessages) >= maxHandledHouseworkMessages {
		handledHouseworkMessages = make(map[string]struct{})
	}
	handledHouseworkMessages[key] = struct{}{}
}

func removeMessageButtons(bot *gotgbot.Bot, cb *gotgbot.CallbackQuery) {
	if cb.Message == nil {
		return
	}
	// A nil slice marshals to null, which Telegram rejects.
	_, _, err := cb.Message.EditReplyMarkup(bot, &gotgbot.EditMessageReplyMarkupOpts{
		ReplyMarkup: gotgbot.InlineKeyboardMarkup{InlineKeyboard: [][]gotgbot.InlineKeyboardButton{}},
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"chat_id":    cb.Message.GetChat().Id,
			"message_id": cb.Message.GetMessageId(),
		}).Warnf("failed to remove buttons from handled message: %s", err.Error())
	}
}

// isAssignee compares the Telegram username with the "@username" stored in the sheet.
func isAssignee(user *gotgbot.User, assignee string) bool {
	if user == nil || user.Username == "" {
		return false
	}
	name := strings.TrimPrefix(strings.TrimSpace(assignee), "@")
	return name != "" && strings.EqualFold(name, user.Username)
}

// nextMemberAfter falls back to the first member when the current assignee is
// not in the member list, so the task never ends up without an assignee.
func nextMemberAfter(members []models.Member, current string) string {
	if len(members) == 0 {
		return current
	}
	for i, member := range members {
		if strings.EqualFold(member.Username, current) {
			return members[(i+1)%len(members)].Username
		}
	}
	return members[0].Username
}

func handleHouseworkAssignToOtherAction(
	bot *gotgbot.Bot,
	ctx *ext.Context,
	housework models.Task,
	numberOfHousework int,
) (callbackAnswer, bool, error) {
	logUserAction(ctx, "housework_assign", fmt.Sprintf("task_id=%d task_name=%s current_assignee=%s", housework.ID, housework.Name, housework.Assignee))

	svc, spreadsheetId, currentSheetName, err := handlers.GetCurrentSheetInfo()
	if err != nil {
		return callbackAnswer{}, false, err
	}

	members, err := handlers.GetMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return callbackAnswer{}, false, err
	}

	currentAssignee := housework.Assignee
	nextAssignee := nextMemberAfter(members, currentAssignee)
	if strings.EqualFold(nextAssignee, currentAssignee) {
		return callbackAnswer{Text: "There is no other member to assign.", Alert: true}, false, nil
	}

	logrus.WithFields(logrus.Fields{
		"user_id":       ctx.EffectiveUser.Id,
		"username":      ctx.EffectiveUser.Username,
		"task_id":       housework.ID,
		"prev_assignee": currentAssignee,
		"next_assignee": nextAssignee,
	}).Info("assigned to other member")

	housework.Assignee = nextAssignee

	taskWeights, err := handlers.GetTaskWeights(svc, spreadsheetId)
	if err != nil {
		logrus.Warnf("failed to get task weights, using default: %s", err.Error())
	}
	housework.TurnsRemaining = getWeightForUser(taskWeights, housework.ID, nextAssignee)

	err = handlers.UpdateHousework(svc, spreadsheetId, currentSheetName, housework, numberOfHousework)
	if err != nil {
		return callbackAnswer{}, false, err
	}

	audit := houseworkAudit{
		Action:       "Reassigned by",
		By:           displayUsername(ctx.EffectiveUser),
		At:           utilities.GetCurrentDateTime(),
		FromAssignee: currentAssignee,
		ToAssignee:   nextAssignee,
	}
	err = sendHouseworkView(bot, ctx, housework, enum.EmojiAssign, "Housework Reassigned", &audit)
	return callbackAnswer{Text: "Assigned to " + nextAssignee}, true, err
}

// getWeightForUser returns the weight for a user on a task, defaulting to 1 if not found
func getWeightForUser(taskWeights map[int]map[string]int, taskID int, username string) int {
	if taskWeights == nil {
		return 1
	}
	if userWeights, ok := taskWeights[taskID]; ok {
		if weight, ok := userWeights[username]; ok && weight > 0 {
			return weight
		}
	}
	return 1
}

func handleHouseworkMarkDoneAction(
	bot *gotgbot.Bot,
	ctx *ext.Context,
	housework models.Task,
	numberOfHousework int,
) (callbackAnswer, bool, error) {
	logUserAction(ctx, "housework_mark_done", fmt.Sprintf("task_id=%d task_name=%s assignee=%s", housework.ID, housework.Name, housework.Assignee))

	// A task without an assignee is open to everyone, otherwise nobody could close it.
	if housework.Assignee != "" && !isAssignee(ctx.EffectiveUser, housework.Assignee) {
		logrus.WithFields(logrus.Fields{
			"user_id":  ctx.EffectiveUser.Id,
			"username": ctx.EffectiveUser.Username,
			"task_id":  housework.ID,
			"assignee": housework.Assignee,
		}).Warn("mark done rejected: user is not the assignee")
		return callbackAnswer{
			Text:  fmt.Sprintf("Only %s can mark \"%s\" as done.", housework.Assignee, housework.Name),
			Alert: true,
		}, false, nil
	}

	svc, spreadsheetId, currentSheetName, err := handlers.GetCurrentSheetInfo()
	if err != nil {
		return callbackAnswer{}, false, err
	}

	taskWeights, err := handlers.GetTaskWeights(svc, spreadsheetId)
	if err != nil {
		logrus.Errorf("failed to get task weights: %s", err.Error())
		taskWeights = nil
	}

	members, err := handlers.GetMembers(svc, spreadsheetId, currentSheetName)
	if err != nil {
		return callbackAnswer{}, false, err
	}

	currentAssignee := housework.Assignee
	if housework.TurnsRemaining > 1 {
		housework.TurnsRemaining--

		logrus.WithFields(logrus.Fields{
			"user_id":         ctx.EffectiveUser.Id,
			"task_id":         housework.ID,
			"assignee":        currentAssignee,
			"turns_remaining": housework.TurnsRemaining,
		}).Info("decremented turns remaining, keeping same assignee")
	} else {
		nextAssignee := nextMemberAfter(members, currentAssignee)
		nextWeight := getWeightForUser(taskWeights, housework.ID, nextAssignee)
		housework.Assignee = nextAssignee
		housework.TurnsRemaining = nextWeight

		logrus.WithFields(logrus.Fields{
			"user_id":         ctx.EffectiveUser.Id,
			"task_id":         housework.ID,
			"prev_assignee":   currentAssignee,
			"next_assignee":   nextAssignee,
			"turns_remaining": nextWeight,
		}).Info("rotated to next assignee with weighted turns")
	}

	housework.LastDone = utilities.GetCurrentDate()
	nextDue, err := utilities.AddDay(housework.LastDone, housework.Frequency)
	if err != nil {
		logrus.Errorf("failed to add day: %s", err.Error())
		return callbackAnswer{}, false, err
	}
	housework.NextDue = nextDue

	err = handlers.UpdateHousework(svc, spreadsheetId, currentSheetName, housework, numberOfHousework)
	if err != nil {
		return callbackAnswer{}, false, err
	}

	audit := houseworkAudit{
		Action:       "Done by",
		By:           displayUsername(ctx.EffectiveUser),
		At:           utilities.GetCurrentDateTime(),
		FromAssignee: currentAssignee,
		ToAssignee:   housework.Assignee,
	}
	err = sendHouseworkView(bot, ctx, housework, enum.EmojiCheck, "Housework Done", &audit)
	return callbackAnswer{Text: "Marked as done"}, true, err
}

func sendHouseworkView(
	bot *gotgbot.Bot,
	ctx *ext.Context,
	housework models.Task,
	icon string,
	title string,
	audit *houseworkAudit,
) error {
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         enum.EmojiCheck + " Mark as done",
					CallbackData: fmt.Sprintf("housework.%d.done", housework.ID),
				},
				{
					Text:         enum.EmojiAssign + " Assign to other",
					CallbackData: fmt.Sprintf("housework.%d.assign", housework.ID),
				},
			},
			{
				{Text: enum.EmojiBroom + " All tasks", CallbackData: HouseworkListCommand},
			},
		},
	}

	text := fmt.Sprintf("%s *%s*\n\n", icon, title)
	if audit != nil {
		text += formatHouseworkAudit(*audit) + "\n---\n"
	}
	text += handlers.ConvertHouseworkToMarkdownFormat(housework)

	_, err := ctx.EffectiveMessage.Reply(
		bot,
		text,
		&gotgbot.SendMessageOpts{
			ReplyMarkup: inlineKeyboard,
			ParseMode:   "markdown",
		},
	)
	return err
}

func formatHouseworkAudit(audit houseworkAudit) string {
	transition := fmt.Sprintf("%s -> %s", audit.FromAssignee, audit.ToAssignee)
	if audit.FromAssignee == audit.ToAssignee {
		transition = fmt.Sprintf("%s (no change)", audit.ToAssignee)
	}
	return fmt.Sprintf(
		"%s %s: %s\n%s At: %s\n%s Assignee: %s",
		enum.EmojiUser, audit.Action, audit.By,
		enum.EmojiClock, audit.At,
		enum.EmojiAssign, transition,
	)
}

// NotifyDueTasks sends a notification to the channel when there are tasks due today or overdue.
func NotifyDueTasks(bot *gotgbot.Bot) {
	if !IsReminderEnabled() {
		logrus.Debug("housework reminders are disabled, skipping notification")
		return
	}

	houseworkMap, err := handlers.GetHouseworkMap()
	if err != nil {
		logrus.Errorf("failed to get housework map: %s", err.Error())
		return
	}

	for _, task := range sortedTasks(houseworkMap) {
		isDateDueOrOverdue, _ := utilities.IsDateDueOrOverdue(task.NextDue)
		if !isDateDueOrOverdue {
			continue
		}

		dueDate := task.NextDue
		if status := handlers.DueStatusText(task.NextDue); status != "" {
			dueDate = fmt.Sprintf("%s (%s)", task.NextDue, status)
		}
		message := fmt.Sprintf(
			"%s *Reminder: %s is due*\n\n"+
				"%s Assignee: %s\n"+
				"%s Due date: %s",
			enum.EmojiBell, task.Name,
			enum.EmojiUser, task.Assignee,
			enum.EmojiCalendar, dueDate,
		)

		_, err = bot.SendMessage(
			task.ChannelId,
			message,
			&gotgbot.SendMessageOpts{
				ParseMode: "markdown",
				ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
					InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
						{
							{
								Text:         enum.EmojiCheck + " Mark as done",
								CallbackData: fmt.Sprintf("housework.%d.done", task.ID),
							},
							{
								Text:         enum.EmojiDetails + " View details",
								CallbackData: fmt.Sprintf("housework.%d.view", task.ID),
							},
						},
					},
				},
			},
		)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"channel_id": task.ChannelId,
				"task_id":    task.ID,
				"task_name":  task.Name,
			}).Errorf("failed to send notification: %s", err.Error())
		} else {
			logrus.WithFields(logrus.Fields{
				"channel_id": task.ChannelId,
				"task_id":    task.ID,
				"task_name":  task.Name,
				"assignee":   task.Assignee,
			}).Info("sent due task notification")
		}
	}
}
