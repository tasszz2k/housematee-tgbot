package main

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/callbackquery"
	"log"
	"os"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
)

// This bot demonstrates some example interactions with commands on telegram.
// It has a basic start command with a bot intro.
// It also has a source command, which sends the bot sourcecode, as a file.
func main() {
	// Get token from the environment variable
	token := os.Getenv("TOKEN")
	if token == "" {
		panic("TOKEN environment variable is empty")
	}

	// Create bot from environment value.
	b, err := gotgbot.NewBot(token, nil)
	if err != nil {
		panic("failed to create new bot: " + err.Error())
	}

	// Create updater and dispatcher.
	updater := ext.NewUpdater(&ext.UpdaterOpts{
		Dispatcher: ext.NewDispatcher(&ext.DispatcherOpts{
			// If an error is returned by a handler, log it and continue going.
			Error: func(b *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
				log.Println("an error occurred while handling update:", err.Error())
				return ext.DispatcherActionNoop
			},
			MaxRoutines: ext.DefaultMaxRoutines,
		}),
	})
	dispatcher := updater.Dispatcher

	// /start command to introduce the bot
	dispatcher.AddHandler(handlers.NewCommand("start", start))
	// /source command to send the bot source code
	dispatcher.AddHandler(handlers.NewCommand("help", help))

	dispatcher.AddHandler(handlers.NewCommand("splitbill", splitBill))
	dispatcher.AddHandler(handlers.NewCommand("remindhousework", remindHousework))
	dispatcher.AddHandler(handlers.NewCommand("settings", settings))
	dispatcher.AddHandler(handlers.NewCommand("feedback", feedback))

	dispatcher.AddHandler(handlers.NewCallback(callbackquery.Prefix("help"), handleButtonCallback))

	// Start receiving updates.
	err = updater.StartPolling(b, &ext.PollingOpts{
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
	log.Printf("%s has been started...\n", b.User.Username)

	// Idle, to keep updates coming in, and avoid bot stopping.
	updater.Idle()
}

func help(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("/help called")
	// show buttons for these commands
	// 	/splitbill - Easily split and manage home bills among your housemates.
	//	/remindhousework - Set reminders for housework tasks and keep your living space clean and organized.
	//	/help - Get assistance on how to use Housematee.
	//	/settings - Customize Housematee to suit your preferences.
	//	/feedback - Share your feedback and suggestions with us to improve Housematee.

	log.Println("/help called")

	// Create an inline keyboard with buttons for each command
	inlineKeyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{Text: "/splitbill", CallbackData: "help.splitbill"},
				{Text: "/remindhousework", CallbackData: "help.remindhousework"},
			},
			{
				{Text: "/settings", CallbackData: "help.settings"},
				{Text: "/feedback", CallbackData: "help.feedback"},
			},
		},
	}

	// Reply to the user with the available commands as buttons
	_, err := ctx.EffectiveMessage.Reply(b, "Here are the available commands:", &gotgbot.SendMessageOpts{
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		return fmt.Errorf("failed to send /help response: %w", err)
	}
	return nil
}

// start introduces the bot.
func start(b *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("/start called")
	_, err := ctx.EffectiveMessage.Reply(b, fmt.Sprintf("Hello, I'm @%s. I <b>repeat</b> all your messages.", b.User.Username), &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}

func splitBill(b *gotgbot.Bot, ctx *ext.Context) error {
	// Sample logic for /splitbill command
	// You can customize this logic based on your requirements
	_, err := ctx.EffectiveMessage.Reply(b, "To split bills, you can use a shared spreadsheet or a bill-splitting app to keep track of expenses and settle bills among your housemates.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send /splitbill response: %w", err)
	}
	return nil
}

func remindHousework(b *gotgbot.Bot, ctx *ext.Context) error {
	// Sample logic for /remindhousework command
	// You can customize this logic based on your requirements
	_, err := ctx.EffectiveMessage.Reply(b, "You can set up housework reminders using a shared task management app or a physical chore calendar to ensure everyone contributes to keeping the living space clean and organized.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send /remindhousework response: %w", err)
	}
	return nil
}

func settings(b *gotgbot.Bot, ctx *ext.Context) error {
	// Sample logic for /settings command
	// You can customize this logic based on your requirements
	_, err := ctx.EffectiveMessage.Reply(b, "Settings customization is not yet implemented in this version of the bot. Stay tuned for future updates!", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send /settings response: %w", err)
	}
	return nil
}

func feedback(b *gotgbot.Bot, ctx *ext.Context) error {
	// Sample logic for /feedback command
	// You can customize this logic based on your requirements
	_, err := ctx.EffectiveMessage.Reply(b, "We appreciate your feedback! Please send your suggestions and feedback to [your-email@example.com](mailto:your-email@example.com) to help us improve Housematee.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send /feedback response: %w", err)
	}
	return nil
}

func handleButtonCallback(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	// Check the CallbackData to determine which button was clicked
	switch cb.Data {
	case "help.splitbill":
		// Handle the /splitbill button click
		err := handleSplitBill(b, ctx)
		if err != nil {
			return err
		}
	case "help.remindhousework":
		// Handle the /remindhousework button click
		err := handleRemindHousework(b, ctx)
		if err != nil {
			return err
		}
	case "help.settings":
		// Handle the /settings button click
		err := handleSettings(b, ctx)
		if err != nil {
			return err
		}
	case "help.feedback":
		// Handle the /feedback button click
		err := handleFeedback(b, ctx)
		if err != nil {
			return err
		}
	default:
		// Handle other button clicks (if any)
	}

	// Send a response to acknowledge the button click
	_, err := cb.Answer(b, &gotgbot.AnswerCallbackQueryOpts{
		Text: "You pressed a button!",
	})
	if err != nil {
		return fmt.Errorf("failed to answer callback query: %w", err)
	}

	return nil
}

// Implement the logic for handling /splitbill button click
func handleSplitBill(b *gotgbot.Bot, ctx *ext.Context) error {
	// Your logic for /splitbill command
	_, err := ctx.EffectiveMessage.Reply(b, "To split bills, you can use a shared spreadsheet or a bill-splitting app to keep track of expenses and settle bills among your housemates.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send /splitbill response: %w", err)
	}
	return nil
}

// Implement the logic for handling /remindhousework button click
func handleRemindHousework(b *gotgbot.Bot, ctx *ext.Context) error {
	// Your logic for /remindhousework command
	_, err := ctx.EffectiveMessage.Reply(b, "You can set up housework reminders using a shared task management app or a physical chore calendar to ensure everyone contributes to keeping the living space clean and organized.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send /remindhousework response: %w", err)
	}
	return nil
}

// Implement the logic for handling /settings button click
func handleSettings(b *gotgbot.Bot, ctx *ext.Context) error {
	// Your logic for /settings command
	_, err := ctx.EffectiveMessage.Reply(b, "Settings customization is not yet implemented in this version of the bot. Stay tuned for future updates!", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send /settings response: %w", err)
	}
	return nil
}

// Implement the logic for handling /feedback button click
func handleFeedback(b *gotgbot.Bot, ctx *ext.Context) error {
	// Your logic for /feedback command
	_, err := ctx.EffectiveMessage.Reply(b, "We appreciate your feedback! Please send your suggestions and feedback to [your-email@example.com](mailto:your-email@example.com) to help us improve Housematee.", &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send /feedback response: %w", err)
	}
	return nil
}
