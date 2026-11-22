package commands

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"housematee-tgbot/enum"
)

func Help(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "help", "command called")

	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: enum.EmojiSplitBill + " " + enum.GetCommandAsText(enum.SplitBillCommand), CallbackData: "help.splitbill"},
				{Text: enum.EmojiHouse + " " + enum.GetCommandAsText(enum.RentCommand), CallbackData: "help.rent"},
			},
			{
				{Text: enum.EmojiBroom + " " + enum.GetCommandAsText(enum.HouseworkCommand), CallbackData: "help.housework"},
				{Text: enum.EmojiSheet + " " + enum.GetCommandAsText(enum.GSheetsCommand), CallbackData: "help.gsheets"},
			},
			{
				{Text: enum.EmojiSettings + " " + enum.GetCommandAsText(enum.SettingsCommand), CallbackData: "help.settings"},
			},
		},
	}

	// Legacy Markdown treats "_" as italics, so commands with underscores are left out of this text.
	message := fmt.Sprintf(
		"%s *Help*\n\n"+
			"%s /splitbill - Add and track shared expenses\n"+
			"%s /rent - Add the monthly rent and bills\n"+
			"%s /housework - See chores and who is next\n"+
			"%s /hw1, /hw2 - Mark a chore as done\n"+
			"%s /gsheets - Create the sheet for a new month\n"+
			"%s /settings - Turn reminders on or off\n"+
			"%s /cancel - Stop the current step\n\n"+
			"Tap a button to start:",
		enum.EmojiHelp,
		enum.EmojiSplitBill,
		enum.EmojiHouse,
		enum.EmojiBroom,
		enum.EmojiCheck,
		enum.EmojiSheet,
		enum.EmojiSettings,
		enum.EmojiCross,
	)

	_, err := ctx.EffectiveMessage.Reply(bot, message, &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
		ParseMode:   "markdown",
	})
	if err != nil {
		return fmt.Errorf("failed to send /help response: %w", err)
	}
	return nil
}

func HandleHelpActionCallback(bot *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	logUserAction(ctx, "help_callback", fmt.Sprintf("callback: %s", cb.Data))

	var err error
	switch cb.Data {
	case "help.splitbill":
		err = SplitBill(bot, ctx)
	case "help.rent":
		// Rent runs as a conversation that starts only from the /rent command.
		_, err = ctx.EffectiveMessage.Reply(
			bot,
			enum.EmojiHouse+" *Rent*\n\nSend /rent to start. The bot asks for the total, electric, and water amounts.",
			&gotgbot.SendMessageOpts{ParseMode: "markdown"},
		)
	case "help.housework":
		err = Housework(bot, ctx)
	case "help.gsheets":
		err = GSheets(bot, ctx)
	case "help.settings":
		err = Settings(bot, ctx)
	}
	if err != nil {
		return err
	}

	_, err = cb.Answer(bot, &gotgbot.AnswerCallbackQueryOpts{})
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}
	return nil
}
