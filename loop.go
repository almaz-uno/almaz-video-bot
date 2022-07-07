package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"almaz.uno/dev/almaz-video-bot/pkg/processors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/ryboe/q"
)

// DispatchLoop executes main processing loop for the bot
// https://core.telegram.org/bots/api
func loop(ctx context.Context, token string) error {
	log.Info().Msg("Hello from Almaz video bot")

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

func processUpdate(ctx context.Context, botAPI *tgbotapi.BotAPI, update tgbotapi.Update) {
	lg := log.With().Int("updateID", update.UpdateID).Logger()
	lg.Debug().Msg("Update processing...")

	if update.Message != nil && update.Message.Text == "/stop" {
		if p, e := os.FindProcess(os.Getpid()); e == nil {
			p.Signal(os.Interrupt)
			return
		} else {
			lg.Warn().Err(e).Msg("Unable to stop current process")
		}
	}

	for _, msg := range processors.Do(ctx, lg, update) {
		if _, e := botAPI.Send(msg); e != nil {
			lg.Warn().Err(e).Msg("Error while processing update")
		}
	}
	lg.Info().Msg("Update processed")
}
