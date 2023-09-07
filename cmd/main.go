package main

import (
	"context"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/conversation"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"housematee-tgbot/commands"
	"housematee-tgbot/config"
	"housematee-tgbot/enum"
	services "housematee-tgbot/services/gsheets"
	"log"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

func main() {
	// Load configuration
	config.Load()
	// init service
	_, err := services.InitGSheetsSvc(context.TODO(), config.GetAppConfig().GoogleApis.Credentials)
	if err != nil {
		panic("failed to init google sheets service: " + err.Error())
	}

	initTelegramBot()
}

func initTelegramBot() {
	// Get token from the environment variable
	token := config.GetAppConfig().Telegram.ApiToken
	if token == "" {
		panic("TOKEN environment variable is empty")
	}

	// Create bot from environment value.
	bot, err := gotgbot.NewBot(token, nil)
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	// Create updater and dispatcher.
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		Dispatcher: ext.NewDispatcher(&ext.DispatcherOpts{
			// If a handler returns an error, log it and continue going.
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				log.Println("an error occurred while handling update:", err.Error())
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		}),
	})
	dispatcher := updater.Dispatcher

	// handle commands
	registerCommandHandlers(dispatcher)

	// Start receiving updates.
	err = updater.StartPolling(bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 9,
			RequestOpts: &gotgbot.RequestOpts{
				Timeout: time.Second * 10,
			},
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	log.Printf("%s has been started...\n", bot.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

// registerCommandHandlers registers all the command handlers for the bot.
// - Supported commands:
//   - /hello - A greeting command to initiate interaction with the bot.
//   - /gsheets - Manage and interact with your Google Sheets data directly from the bot.
//   - /splitbill - Easily split expenses with your housemates and keep track of who owes what.
//   - /housework - Organize and delegate house chores among housemates with reminders and schedules.
//   - /settings - Adjust bot settings, such as language, notification preferences, and more.
//   - /feedback - Provide feedback about the bot or report issues for continuous improvement.
//   - /help - Get a list of available commands and learn how to use the bot effectively.
func registerCommandHandlers(dispatcher *ext.Dispatcher) {

	// Register commands handlers
	dispatcher.AddHandler(handlers.NewCommand(enum.HelloCommand, commands.Hello))
	dispatcher.AddHandler(handlers.NewCommand(enum.SplitBillCommand, commands.SplitBill))
	dispatcher.AddHandler(handlers.NewCommand(enum.HouseworkCommand, commands.Housework))
	dispatcher.AddHandler(handlers.NewCommand(enum.HelpCommand, commands.Help))
	dispatcher.AddHandler(handlers.NewCommand(enum.SettingsCommand, commands.Settings))
	dispatcher.AddHandler(handlers.NewCommand(enum.FeedbackCommand, commands.Feedback))
	dispatcher.AddHandler(handlers.NewCommand(enum.GSheetsCommand, commands.GSheets))

	// Register conversation handlers
	// Register conversation handlers for the split bill
	dispatcher.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCallback(callbackquery.Equal("splitbill.add"), commands.StartAddSplitBill)},
		map[string][]ext.Handler{
			enum.AddExpense: {handlers.NewMessage(commands.NoCommands, commands.AddExpenseConversationHandler)},
		},
		&handlers.ConversationOpts{
			Exits:        []ext.Handler{handlers.NewCommand(enum.CancelCommand, commands.Cancel)},
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
			AllowReEntry: true,
		},
	))

	// Register callback query handlers
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("help."), commands.HandleHelpActionCallback))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("splitbill."), commands.HandleSplitBillActionCallback))
	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("housework."), commands.HandleHouseworkActionCallback))

	// Register conversation handlers

}
