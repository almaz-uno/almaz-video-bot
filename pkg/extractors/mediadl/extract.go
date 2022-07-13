package mediadl

import (
	"context"
	"net/url"
	"regexp"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"github.com/ryboe/q"
)

type (
	Extractor struct {
		botAPI    *tgbotapi.BotAPI
		wd        string
		urlPrefix string
		me        int64
	}
)

func NewExtractor(botAPI *tgbotapi.BotAPI, wd, urlPrefix string, me int64) *Extractor {
	return &Extractor{
		botAPI:    botAPI,
		wd:        wd,
		urlPrefix: urlPrefix,
		me:        me,
	}
}

var (
	regexpURL    = regexp.MustCompile(`(https?):\/\/([\w_-]+(?:(?:\.[\w_-]+)+))([\w.,@?^=%&:\/~+#-]*[\w@?^=%&\/~+#-])`)
	htmlReplacer = strings.NewReplacer(
		`<`, `&lt;`,
		`>`, `&gt;`,
		`&`, `&amp;`,
	)
)

// getLinks returns all links from text
func getLinks(ctx context.Context, text string) []string {
	return regexpURL.FindAllString(text, -1)
}

// Extract extracts and print
func (extractor *Extractor) Extract(ctx context.Context, update *tgbotapi.Update) {
	lg := log.With().Int("update", update.UpdateID).Logger()

	// callbackPrefix := extractor.me + "⇒"
	formats := []string{VideoFormat, AudioOnlyFormat}

	var message *tgbotapi.Message

	switch {
	case update.Message != nil && update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == extractor.me:
		// new quality request!!!
		formats = []string{strings.TrimSpace(update.Message.Text)}
		message = update.Message.ReplyToMessage
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

	lg = lg.With().Int64("chat-id", chatID).Int64("reply-to-id", int64(replyTo)).Logger()

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
		for _, f := range formats {
			info, di, e := YtDlp(ctx, extractor.wd, f, u)

			q.Q(di.Command())

			lgg := lgu.With().
				Str("url", u).
				Str("format", f).
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
		mk := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				kbb...,
			),
		)

		mc := tgbotapi.NewMessage(chatID, `<a href="`+u+`">🎥 ⇒ `+htmlReplacer.Replace(title)+`</a>`)
		mc.ReplyToMessageID = replyTo
		mc.ParseMode = "HTML"
		mc.DisableWebPagePreview = true
		mc.ReplyMarkup = mk

		if _, e := extractor.botAPI.Send(mc); e != nil {
			lgu.Error().Err(e).Msg("Unable to send media")
		}

	}
}
