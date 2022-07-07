package processors

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/ryboe/q"
)

func Do(ctx context.Context, log zerolog.Logger, update tgbotapi.Update) []tgbotapi.Chattable {
	var (
		output  []tgbotapi.Chattable
		message *tgbotapi.Message
	)

	switch {
	case update.Message != nil:
		message = update.Message
	case update.EditedMessage != nil:
		message = update.EditedMessage
	}

	if message == nil {
		return output
	}

	output = append(output, zen(ctx, log, message)...)
	return output
}

var zenURL = regexp.MustCompilePOSIX(`https://zen.yandex.ru/video/watch/[[:alnum:]]*`)

func zen(ctx context.Context, log zerolog.Logger, message *tgbotapi.Message) []tgbotapi.Chattable {
	var output []tgbotapi.Chattable
	chatID := message.Chat.ID

	text := message.Text

	URLs := zenURL.FindAllString(text, -1)
	q.Q(text, URLs)

	for _, u := range URLs {
		if m3u8s, e := extractM3U8(ctx, u); e != nil {
			log.Warn().Str("url", u).Msg("Unable to extract m3u8")
		} else {
			for _, m := range m3u8s {
				mc := tgbotapi.NewMessage(chatID, m)

				mc.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonURL("m3u8", m),
					),
				)

				output = append(output, mc)
			}
		}
	}

	return output
}

var m3u8URL = regexp.MustCompilePOSIX(`(http|https)://[a-zA-Z0-9./?&,=_-]*?`)

func extractM3U8(ctx context.Context, URL string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	bb, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	text := string(bb)

	URLs := m3u8URL.FindAllString(text, -1)
	var result []string

	for _, m := range URLs {
		if strings.Contains(m, "master.m3u8") {
			// result = append(result, m)
			return []string{m}, nil
		}
	}
	return result, nil
}
