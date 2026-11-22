package commands

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"housematee-tgbot/enum"
)

func Feedback(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "feedback", "command called")
	message := enum.EmojiMail + " *Feedback*\n\n" +
		"Thank you for helping us make the bot better.\n\n" +
		"Send your ideas or bug reports to:\n" +
		"[tasszz2k@gmail.com](mailto:tasszz2k@gmail.com)"
	_, err := ctx.EffectiveMessage.Reply(bot, message, &gotgbot.SendMessageOpts{
		ParseMode: "markdown",
	})
	if err != nil {
		return fmt.Errorf("failed to send /feedback response: %w", err)
	}
	return nil
}
