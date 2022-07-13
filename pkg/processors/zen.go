package processors

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/ryboe/q"
	"golang.org/x/net/html"
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
		if m3u8, title, e := extractM3U8(ctx, u); e != nil {
			log.Warn().Str("url", u).Msg("Unable to extract m3u8")
		} else {
			if title == "" {
				title = "zen video"
			}

			//			tgbotapi.NewVideo(chatID)

			mc := tgbotapi.NewMessage(chatID, m3u8)
			mc.ReplyToMessageID = message.MessageID
			mc.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL(title, m3u8),
				),
			)
			output = append(output, mc)

		}
	}

	return output
}

var (
	m3u8URL = regexp.MustCompilePOSIX(`(http|https)://[a-zA-Z0-9./?&,=_-]*?`)
	titleRe = regexp.MustCompilePOSIX(`<title></title> ()://[a-zA-Z0-9./?&,=_-]*?`)
)

func extractM3U8(ctx context.Context, URL string) (string, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, URL, nil)
	if err != nil {
		return "", "", err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()

	title := ""

	bb, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}
	text := string(bb)

	if doc, err := html.Parse(bytes.NewBuffer(bb)); err == nil {
		var f func(*html.Node)
		f = func(n *html.Node) {
			if n.Type == html.ElementNode && n.Data == "title" {
				title = n.FirstChild.Data
				return
			}
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				f(c)
			}
		}
		f(doc)
	}

	URLs := m3u8URL.FindAllString(text, -1)

	for _, m := range URLs {
		if strings.Contains(m, "master.m3u8") {
			return m, title, nil
		}
	}
	return "", title, nil
}
