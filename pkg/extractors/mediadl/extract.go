package mediadl

import (
	"context"
	"net/url"
	"regexp"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/ryboe/q"
)

type (
	Extractor struct {
		botAPI    *tgbotapi.BotAPI
		wd        string
		urlPrefix string
	}
)

func NewExtractor(botAPI *tgbotapi.BotAPI, wd, urlPrefix string) *Extractor {
	return &Extractor{
		wd:        wd,
		botAPI:    botAPI,
		urlPrefix: urlPrefix,
	}
}

var regexpURL = regexp.MustCompile(`(https?):\/\/([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:\/~+#-]*[\w@?^=%&\/~+#-])`)

// getLinks returns all links from text
func getLinks(ctx context.Context, text string) []string {
	return regexpURL.FindAllString(text, -1)
}

// Extract extracts and print
func (extractor *Extractor) Extract(ctx context.Context, update *tgbotapi.Update) {
	lg := log.With().Int("update", update.UpdateID).Logger()

	var message *tgbotapi.Message

	switch {
	case update.Message != nil:
		message = update.Message
	case update.EditedMessage != nil:
		message = update.EditedMessage
	default:
		lg.Info().Msg("Ignore this update")
		return
	}

	// from := message.From.ID
	text := message.Text
	replyTo := message.MessageID
	chatID := message.Chat.ID

	URLs := regexpURL.FindAllString(text, -1)

	for _, e := range message.CaptionEntities {
		if e.URL != "" {
			URLs = append(URLs, e.URL)
		}
	}

	for _, e := range message.Entities {
		if e.URL != "" {
			URLs = append(URLs, e.URL)
		}
	}

	for _, u := range URLs {
		lgu := lg.With().Str("url", u).Logger()
		title := "<media>"
		kbb := make([]tgbotapi.InlineKeyboardButton, 0)
		for _, f := range []func(ctx context.Context, dir, URL string) (*YtDlpInfo, []byte, error){DownloadVideo, DownloadAudio} {
			info, ytDlpRaw, e := f(ctx, extractor.wd, u)
			lgg := lgu.With().
				Str("url", u).
				Str("format-id", info.FormatID).
				Logger()

			lgg.Debug().Msg("Processing URL...")

			if e != nil {
				lgg.Error().Err(e).Msg("Unable to download")
				q.Q(e)
				continue
			} else {
				lgg.Info().Err(e).Msg("Successfully downloaded")
			}

			if len(info.RequestedDownloads) == 0 {
				lgg.Error().Msg("len(info.RequestedDownloads) == 0")
				q.Q(string(ytDlpRaw))
				continue
			}
			title = info.Title
			for _, rd := range info.RequestedDownloads {
				kbb = append(kbb, tgbotapi.NewInlineKeyboardButtonURL(
					rd.Ext+": "+rd.Resolution,
					extractor.urlPrefix+url.PathEscape(rd.Filename)))
			}

		}

		if len(kbb) == 0 {
			continue
		}
		mc := tgbotapi.NewMessage(chatID, "🎥 ⇒ "+title)
		mc.ReplyToMessageID = replyTo
		mc.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				kbb...,
			),
		)

		if _, e := extractor.botAPI.Send(mc); e != nil {
			lgu.Error().Err(e).Msg("Unable to send media")
		}

	}
}
