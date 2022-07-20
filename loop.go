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
	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return fmt.Errorf("unable to acquire Telegram bot API: %w", err)
	}
	botAPI.Debug = false

	log.Info().Msgf("Authorized on account %s", botAPI.Self.UserName)

	extractor := mediadl.NewExtractor(botAPI, cfgMediaDir, cfgServerPrefix+cfgStaticPrefix, botAPI.Self.ID)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := botAPI.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			log.Info().Msg("Bye!")
			return ctx.Err()
		case update := <-updates:
			bb, _ := json.MarshalIndent(update, "  ", "  ")
			q.Q("raw update is", string(bb))
			go processUpdate(ctx, botAPI, update, extractor)
		}
	}
}

const announceText = `
To get video and audio from supported resources, please send video or audio link to this bot.
Bot will extracts multimedia URLs from message if any. It will provide appropriate possibilities for media downloading.

Downloading provided by great tool https://github.com/yt-dlp/yt-dlp.

You can inspect source codes of this bot in: https://github.com/almaz-uno/almaz-video-bot
`

func processUpdate(ctx context.Context, botAPI *tgbotapi.BotAPI, update tgbotapi.Update, extractor *mediadl.Extractor) {
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
		if _, e := botAPI.Send(tgbotapi.NewMessage(update.Message.Chat.ID, announceText)); e != nil {
			lg.Error().Err(e).Msg("Error while sending announce message")
		}
	default:
		extractor.ProcessUpdate(ctx, &update)
	}
	lg.Info().Msg("Update processed")
}
