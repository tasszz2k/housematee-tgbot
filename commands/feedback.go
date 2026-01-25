package commands

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func Feedback(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "feedback", "command called")
	message := "*Feedback*\n\n" +
		"We appreciate your feedback!\n\n" +
		"Please send your suggestions to:\n" +
		"[tasszz2k@gmail.com](mailto:tasszz2k@gmail.com)"
	_, err := ctx.EffectiveMessage.Reply(bot, message, &gotgbot.SendMessageOpts{
		ParseMode: "markdown",
	})
	if err != nil {
		return fmt.Errorf("failed to send /feedback response: %w", err)
	}
	return nil
}
