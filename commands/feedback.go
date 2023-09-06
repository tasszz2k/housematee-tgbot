package commands

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func Feedback(bot *gotgbot.Bot, ctx *ext.Context) error {
	// Sample logic for /feedback command
	// You can customize this logic based on your requirements
	_, err := ctx.EffectiveMessage.Reply(bot, "We appreciate your feedback! Please send your suggestions and feedback to [tasszz2k@gmail.com](mailto:tasszz2k@gmail.com) to help us improve Housematee.", &gotgbot.SendMessageOpts{
		ParseMode: "markdown",
	})
	if err != nil {
		return fmt.Errorf("failed to send /feedback response: %w", err)
	}
	return nil
}
