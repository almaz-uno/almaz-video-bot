package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"almaz.uno/dev/almaz-video-bot/pkg/extractors/mediadl"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/ryboe/q"
)

// DispatchLoop executes main processing loop for the bot
// https://core.telegram.org/bots/api
func loop(ctx context.Context, token string) error {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("unable to acquire Telegram bot API: %w", err)
	}
	bot.Debug = false

	log.Info().Msgf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Bye!")
			return ctx.Err()
		case update := <-updates:
			bb, _ := json.MarshalIndent(update, "  ", "  ")
			q.Q("raw update is", string(bb))
			go processUpdate(ctx, bot, update)
		}
	}
}

const announceText = `
To get video and audio from supported resources, please send video link to the this bot.
Bot will download video (and audio if supported) with https://github.com/yt-dlp/yt-dlp and return link to this.
`

func processUpdate(ctx context.Context, botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	lg := log.With().Int("updateID", update.UpdateID).Logger()
	lg.Debug().Msg("Update processing...")

	switch {
	case update.Message != nil && update.Message.Text == "/stop":
		if p, e := os.FindProcess(os.Getpid()); e == nil {
			p.Signal(os.Interrupt)
			return
		} else {
			lg.Warn().Err(e).Msg("Unable to stop current process")
		}
	case update.Message != nil && update.Message.Text == "/start":
		if _, err := botAPI.Send(tgbotapi.NewMessage(update.Message.Chat.ID, announceText)); err != nil {
			lg.Error().Err(err).Msg("Error while sending announce message")
		}
	default:
		mediadl.NewExtractor(botAPI, cfgMediaDir, cfgServerPrefix+cfgStaticPrefix).Extract(ctx, &update)
	}
	lg.Info().Msg("Update processed")
}
