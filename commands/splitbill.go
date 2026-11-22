package commands

import (
	"fmt"
	"strconv"
	"strings"
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

// pendingUpdateExpense stores the expense being updated (keyed by user ID)
var (
	pendingUpdateExpense = make(map[int64]*models.Expense)
	pendingUpdateMutex   sync.RWMutex
)

// SplitBill handles the /splitbill command.
func SplitBill(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "splitbill", "command called")
	// show buttons for these commands
	// - Supported commands:
	// - /add - Add a new expense to the bill.
	// - /view - show last 5 records as table.
	// - /update - update a record.
	// - /delete - delete a record.
	// - /report - show report.

	// Create an inline keyboard with buttons for each command
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: enum.EmojiAdd + " Add", CallbackData: "splitbill.add"},
				{Text: enum.EmojiView + " View", CallbackData: "splitbill.view"},
			},
			{
				{Text: enum.EmojiEdit + " Update", CallbackData: "splitbill.update"},
				{Text: enum.EmojiDelete + " Delete", CallbackData: "splitbill.delete"},
			},
			{
				{Text: enum.EmojiReport + " Show Report", CallbackData: "splitbill.report"},
			},
		},
	}

	// Reply to the user with the available commands as buttons
	_, err := ctx.EffectiveMessage.Reply(
		bot,
		enum.EmojiSplitBill+" *Split Bill*\n\nChoose an action:",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: inlineKeyboard,
			ParseMode:   "markdown",
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send /splitbill response: %w", err)
	}
	return nil
}

func HandleSplitBillActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	logUserAction(ctx, "splitbill_callback", fmt.Sprintf("callback: %s", cb.Data))

	var err error

	// Check the CallbackData to determine which button was clicked
	switch cb.Data {
	case "splitbill.add":
		err = StartAddSplitBill(bot, ctx)
	case "splitbill.view":
		err = HandleSplitBillViewActionCallback(bot, ctx)
	case "splitbill.update":
		err = HandleSplitBillUpdateAction(bot, ctx)
	case "splitbill.delete":
		err = HandleSplitBillDeleteAction(bot, ctx)
	case "splitbill.report":
		err = HandleSplitBillReportActionCallback(bot, ctx)
	case "splitbill.back":
		err = HandleSplitBillBack(bot, ctx)
	case "splitbill.delete.cancel":
		err = HandleCancelDelete(bot, ctx)
	default:
		// Handle dynamic callbacks with IDs
		if strings.HasPrefix(cb.Data, "splitbill.update.") && !strings.Contains(cb.Data, "confirm") {
			err = HandleSelectExpenseForUpdate(bot, ctx)
		} else if strings.HasPrefix(cb.Data, "splitbill.delete.confirm.") {
			err = HandleConfirmDelete(bot, ctx)
		} else if strings.HasPrefix(cb.Data, "splitbill.delete.") && !strings.Contains(cb.Data, "cancel") {
			err = HandleSelectExpenseForDelete(bot, ctx)
		}
	}

	if err != nil {
		return err
	}

	// Send a response to acknowledge the button click
	_, err = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{})
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	return nil
}

func HandleSplitBillViewActionCallback(
	bot *gotgbot.Bot,
	ctx *ext.Context,
) error {
	return handlers.HandleSplitBillViewAction(bot, ctx)
}

func HandleSplitBillReportActionCallback(
	bot *gotgbot.Bot,
	ctx *ext.Context,
) error {
	return handlers.HandleSplitBillReportAction(bot, ctx)
}

func StartAddSplitBill(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "splitbill_add", "starting add expense flow")
	// Prompt the user to enter the details
	htlmText := fmt.Sprintf(
		`%s <b>Add Expense</b>

Send the details, one item on each line:
<code>name
amount
date</code> <i>(optional, default: %s)</i>
<code>payer</code> <i>(optional, default: %s)</i>

Example:
<code>Groceries
150k</code>

Amounts accept k (thousand) and m (million). Send /cancel to stop.`,
		enum.EmojiAdd,
		utilities.GetCurrentDate(),
		escapeHTML(displayUsername(ctx.EffectiveUser)),
	)
	_, err := ctx.EffectiveMessage.Reply(
		bot, htlmText, &gotgbot.SendMessageOpts{
			ParseMode: "html",
		},
	)
	if err != nil {
		return err
	}
	return tgBotHandler.NextConversationState(enum.AddExpense)
}

func AddExpenseConversationHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	if !CheckPermission(bot, ctx) {
		return nil
	}
	return handlers.HandleExpenseAddAction(bot, ctx)
}

// ==================== UPDATE FLOW ====================

// HandleSplitBillUpdateAction shows list of recent expenses to select for update
func HandleSplitBillUpdateAction(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "splitbill_update", "showing expense list for update")

	expenses, err := handlers.GetRecentExpenses(5)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("%s *Error*\n\n%s", enum.EmojiCross, err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	if len(expenses) == 0 {
		_, err := ctx.EffectiveMessage.Reply(bot, enum.EmojiReceipt+" *No Expenses*\n\nThere are no expenses to update.", &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	// Build keyboard with expense buttons
	keyboard := make([][]gotgbot.InlineKeyboardButton, 0, len(expenses)+1)
	for _, expense := range expenses {
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
			{Text: expenseButtonLabel(enum.EmojiEdit, expense), CallbackData: fmt.Sprintf("splitbill.update.%d", expense.ID)},
		})
	}
	keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
		{Text: enum.EmojiBack + " Back", CallbackData: "splitbill.back"},
	})

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyboard}

	_, err = ctx.EffectiveMessage.Reply(bot, enum.EmojiEdit+" *Update Expense*\n\nChoose the expense to update:", &gotgbot.SendMessageOpts{
		ParseMode:   "markdown",
		ReplyMarkup: inlineKeyboard,
	})
	return err
}

// HandleSelectExpenseForUpdate shows selected expense and prompts for new values
func HandleSelectExpenseForUpdate(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	// Extract expense ID from callback data: splitbill.update.{id}
	parts := strings.Split(cb.Data, ".")
	if len(parts) < 3 {
		return fmt.Errorf("invalid callback data: %s", cb.Data)
	}

	expenseId, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid expense id: %s", parts[2])
	}

	logUserAction(ctx, "splitbill_update_select", fmt.Sprintf("expense_id=%d", expenseId))

	expense, err := handlers.GetExpenseById(expenseId)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("%s *Error*\n\n%s", enum.EmojiCross, err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	// Store the expense being updated
	pendingUpdateMutex.Lock()
	pendingUpdateExpense[ctx.EffectiveUser.Id] = expense
	pendingUpdateMutex.Unlock()

	// Show current values and prompt for new amount only
	message := fmt.Sprintf(`%s *Update Expense #%d*

%s Name: %s
%s Current amount: %s
%s Date: %s
%s Payer: %s

Send the new amount (for example 150k). Send /cancel to stop.`,
		enum.EmojiEdit, expense.ID,
		enum.EmojiReceipt, expense.Name,
		enum.EmojiMoney, expense.Amount,
		enum.EmojiCalendar, expense.Date,
		enum.EmojiUser, expense.Payer,
	)

	_, err = ctx.EffectiveMessage.Reply(bot, message, &gotgbot.SendMessageOpts{
		ParseMode: "markdown",
	})
	if err != nil {
		return err
	}

	return tgBotHandler.NextConversationState(enum.UpdateExpense)
}

// UpdateExpenseConversationHandler processes user input and updates the expense
func UpdateExpenseConversationHandler(bot *gotgbot.Bot, ctx *ext.Context) error {
	if !CheckPermission(bot, ctx) {
		return nil
	}

	// Get the pending expense (this is the old expense)
	pendingUpdateMutex.RLock()
	oldExpense, exists := pendingUpdateExpense[ctx.EffectiveUser.Id]
	pendingUpdateMutex.RUnlock()

	if !exists || oldExpense == nil {
		_, err := ctx.EffectiveMessage.Reply(bot, enum.EmojiCross+" *Error*\n\nNo expense is selected. Start again with /splitbill.", &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	// Create a copy for the new expense
	newExpense := *oldExpense

	// Parse the user's message - amount only
	amountStr := strings.TrimSpace(ctx.EffectiveMessage.Text)
	if amountStr == "" {
		_, err := ctx.EffectiveMessage.Reply(bot, enum.EmojiCross+" *Invalid Amount*\n\nSend a number, for example 150k or 1.2m.", &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}
	newExpense.Amount = utilities.ParseAmount(amountStr)

	// Update in Google Sheets with audit logging
	err := handlers.UpdateExpenseById(*oldExpense, newExpense, displayUsername(ctx.EffectiveUser))
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("%s *Expense Not Updated*\n\n%s", enum.EmojiCross, err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	// Clean up pending update
	pendingUpdateMutex.Lock()
	delete(pendingUpdateExpense, ctx.EffectiveUser.Id)
	pendingUpdateMutex.Unlock()

	// Format amount for display
	newExpense.Amount = utilities.FormatMoney(cast.ToInt(newExpense.Amount))

	// Reply with updated expense and action buttons
	response := enum.EmojiCheck + " *Expense Updated*\n\n" + handlers.FormatExpenseMarkdown(newExpense)

	_, err = ctx.EffectiveMessage.Reply(bot, response, &gotgbot.SendMessageOpts{
		ParseMode:   "markdown",
		ReplyMarkup: handlers.ExpenseActionKeyboard(newExpense.ID),
	})
	if err != nil {
		return err
	}

	return tgBotHandler.EndConversation()
}

func expenseButtonLabel(icon string, expense models.Expense) string {
	return fmt.Sprintf("%s #%d %s (%s)", icon, expense.ID, expense.Name, expense.Amount)
}

// ==================== DELETE FLOW ====================

// HandleSplitBillDeleteAction shows list of recent expenses to select for deletion
func HandleSplitBillDeleteAction(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "splitbill_delete", "showing expense list for delete")

	expenses, err := handlers.GetRecentExpenses(5)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("%s *Error*\n\n%s", enum.EmojiCross, err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	if len(expenses) == 0 {
		_, err := ctx.EffectiveMessage.Reply(bot, enum.EmojiReceipt+" *No Expenses*\n\nThere are no expenses to delete.", &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	// Build keyboard with expense buttons
	keyboard := make([][]gotgbot.InlineKeyboardButton, 0, len(expenses)+1)
	for _, expense := range expenses {
		keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
			{Text: expenseButtonLabel(enum.EmojiDelete, expense), CallbackData: fmt.Sprintf("splitbill.delete.%d", expense.ID)},
		})
	}
	keyboard = append(keyboard, []gotgbot.InlineKeyboardButton{
		{Text: enum.EmojiBack + " Back", CallbackData: "splitbill.back"},
	})

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{InlineKeyboard: keyboard}

	_, err = ctx.EffectiveMessage.Reply(bot, enum.EmojiDelete+" *Delete Expense*\n\nChoose the expense to delete:", &gotgbot.SendMessageOpts{
		ParseMode:   "markdown",
		ReplyMarkup: inlineKeyboard,
	})
	return err
}

// HandleSelectExpenseForDelete shows selected expense with confirm/cancel buttons
func HandleSelectExpenseForDelete(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	// Extract expense ID from callback data: splitbill.delete.{id}
	parts := strings.Split(cb.Data, ".")
	if len(parts) < 3 {
		return fmt.Errorf("invalid callback data: %s", cb.Data)
	}

	expenseId, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid expense id: %s", parts[2])
	}

	logUserAction(ctx, "splitbill_delete_select", fmt.Sprintf("expense_id=%d", expenseId))

	expense, err := handlers.GetExpenseById(expenseId)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("%s *Error*\n\n%s", enum.EmojiCross, err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	// Show expense details with confirm/cancel buttons
	message := fmt.Sprintf(`%s *Delete Expense #%d?*

%s Name: %s
%s Amount: %s
%s Date: %s
%s Payer: %s

Do you want to delete this expense?`,
		enum.EmojiWarning, expense.ID,
		enum.EmojiReceipt, expense.Name,
		enum.EmojiMoney, expense.Amount,
		enum.EmojiCalendar, expense.Date,
		enum.EmojiUser, expense.Payer,
	)

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: enum.EmojiCheck + " Yes, delete", CallbackData: fmt.Sprintf("splitbill.delete.confirm.%d", expenseId)},
				{Text: enum.EmojiCross + " Cancel", CallbackData: "splitbill.delete.cancel"},
			},
		},
	}

	_, err = ctx.EffectiveMessage.Reply(bot, message, &gotgbot.SendMessageOpts{
		ParseMode:   "markdown",
		ReplyMarkup: inlineKeyboard,
	})
	return err
}

// HandleConfirmDelete deletes the expense after user confirmation
func HandleConfirmDelete(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	// Extract expense ID from callback data: splitbill.delete.confirm.{id}
	parts := strings.Split(cb.Data, ".")
	if len(parts) < 4 {
		return fmt.Errorf("invalid callback data: %s", cb.Data)
	}

	expenseId, err := strconv.Atoi(parts[3])
	if err != nil {
		return fmt.Errorf("invalid expense id: %s", parts[3])
	}

	logUserAction(ctx, "splitbill_delete_confirm", fmt.Sprintf("expense_id=%d", expenseId))

	// Fetch the expense first to get name, amount, and existing note
	expense, err := handlers.GetExpenseById(expenseId)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("%s *Error*\n\n%s", enum.EmojiCross, err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	username := displayUsername(ctx.EffectiveUser)

	// Format amount with currency for audit log
	formattedAmount := utilities.FormatMoney(cast.ToInt(expense.Amount))

	// Soft delete the expense (keeps ID, appends deletion to audit log)
	err = handlers.DeleteExpenseById(expenseId, expense.Name, formattedAmount, expense.Note, username)
	if err != nil {
		_, err := ctx.EffectiveMessage.Reply(bot, fmt.Sprintf("%s *Expense Not Deleted*\n\n%s", enum.EmojiCross, err.Error()), &gotgbot.SendMessageOpts{ParseMode: "markdown"})
		return err
	}

	_, err = ctx.EffectiveMessage.Reply(bot, fmt.Sprintf(
		"%s *Expense #%d Deleted*\n\n%s %s (%s)\n%s By: %s\n%s At: %s",
		enum.EmojiDelete, expenseId,
		enum.EmojiReceipt, expense.Name, formattedAmount,
		enum.EmojiUser, username,
		enum.EmojiClock, utilities.GetCurrentDateTime(),
	), &gotgbot.SendMessageOpts{
		ParseMode: "markdown",
	})
	return err
}

// HandleCancelDelete cancels the delete operation
func HandleCancelDelete(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "splitbill_delete_cancel", "delete cancelled")
	_, err := ctx.EffectiveMessage.Reply(bot, enum.EmojiCross+" *Cancelled*\n\nThe expense is not deleted.", &gotgbot.SendMessageOpts{
		ParseMode: "markdown",
	})
	return err
}

// HandleSplitBillBack returns to the main splitbill menu
func HandleSplitBillBack(bot *gotgbot.Bot, ctx *ext.Context) error {
	return SplitBill(bot, ctx)
}
