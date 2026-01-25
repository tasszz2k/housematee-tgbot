package commands

import (
	"fmt"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
)

func Settings(bot *gotgbot.Bot, ctx *ext.Context) error {
	logUserAction(ctx, "settings", "command called")
	_, err := ctx.EffectiveMessage.Reply(bot, "*Settings*\n\nSettings customization is coming soon. Stay tuned for updates!", &gotgbot.SendMessageOpts{
		ParseMode: "markdown",
	})
	if err != nil {
		return fmt.Errorf("failed to send /settings response: %w", err)
	}
	return nil
}
