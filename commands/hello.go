package commands

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"log"
)

// Hello is a simple command that replies to the user with a hello message.
func Hello(bot *gotgbot.Bot, ctx *ext.Context) error {
	log.Println("/hello called")
	textHtml := fmt.Sprintf("Hello <b>@%s</b>, I'm <b>@%s</b>!", ctx.EffectiveUser.Username, bot.User.Username)
	_, err := ctx.EffectiveMessage.Reply(bot, textHtml, &gotgbot.SendMessageOpts{
		ParseMode: "html",
	})
	if err != nil {
		return fmt.Errorf("failed to send start message: %w", err)
	}
	return nil
}
