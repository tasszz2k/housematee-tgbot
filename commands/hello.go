package commands

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"

	"housematee-tgbot/enum"
)

// Hello is a simple command that replies to the user with a hello message.
func Hello(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "hello", "command called")
	textHtml := fmt.Sprintf(
		"%s Hello <b>%s</b>! I am <b>@%s</b>, your housemate helper.\n\nSend /help to see what I can do.",
		enum.EmojiWave,
		escapeHTML(displayUsername(ctx.EffectiveUser)),
		escapeHTML(bot.User.Username),
	)
	_, err := ctx.EffectiveMessage.Reply(bot, textHtml, &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}
