package mediadl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type (
	// Extractor is multithreading updates processor
	Extractor struct {
		botAPI      *tgbotapi.BotAPI
		wd          string
		urlPrefix   string
		linksPrefix string
		me          int64
	}

	botCommand byte

	// button data for the bot
	buttonData string
)

const (
	commandDownload = "download "
	commandInfo     = "info "
	commandFormats  = "formats "
)

const (
	none botCommand = iota
	download
	listFormats
)

func (c buttonData) command() botCommand {
	if len(c) < 1 {
		return none
	}
	return []botCommand(c)[0]
}

func (c buttonData) data() string {
	if len(c) <= 1 {
		return ""
	}
	return string(c)[1:]
}

func newButtonData(command botCommand, data string) buttonData {
	return buttonData(string(command) + data)
}

func NewExtractor(botAPI *tgbotapi.BotAPI, wd, urlPrefix, linksPrefix string, me int64) *Extractor {
	return &Extractor{
		botAPI:      botAPI,
		wd:          wd,
		urlPrefix:   urlPrefix,
		linksPrefix: linksPrefix,
		me:          me,
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

// ProcessUpdate processes update
func (extractor *Extractor) ProcessUpdate(ctx context.Context, update *tgbotapi.Update) {
	lg := log.With().Int("update", update.UpdateID).Logger()

	repliedToMe := update.Message != nil && update.Message.ReplyToMessage != nil && update.Message.ReplyToMessage.From.ID == extractor.me

	callbackToMe := update.CallbackQuery != nil && update.CallbackQuery.Message.From.ID == extractor.me
	cdata := update.CallbackData()

	switch {
	case callbackToMe && strings.HasPrefix(cdata, commandDownload):
		extractor.downloadMedia(ctx, lg, cdata[len(commandDownload):], update.CallbackQuery.Message)
		extractor.botAPI.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, "Download command processed"))
	case callbackToMe && strings.HasPrefix(cdata, commandFormats):
		extractor.showFormats(ctx, lg, update.CallbackQuery.Message)
		extractor.botAPI.Send(tgbotapi.NewCallback(update.CallbackQuery.ID, "Formats command processed"))
	case repliedToMe:
		extractor.downloadMedia(ctx, lg, update.Message.Text, update.Message.ReplyToMessage)
	case update.Message != nil:
		extractor.extractMediaLinks(ctx, lg, update.Message)
	case update.EditedMessage != nil:
		extractor.extractMediaLinks(ctx, lg, update.EditedMessage)
	default:
		lg.Info().Msg("Ignore this update")
		return
	}
}

func extractURLs(message *tgbotapi.Message) []string {
	URLs := regexpURL.FindAllString(message.Text, -1)

	URLs = append(URLs, regexpURL.FindAllString(message.Caption, -1)...)

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

	mu := map[string]bool{}
	for _, u := range URLs {
		mu[u] = true
	}
	URLs = nil
	for u := range mu {
		URLs = append(URLs, u)
	}

	return URLs
}

func (extractor *Extractor) extractMediaLinks(ctx context.Context, lg zerolog.Logger, message *tgbotapi.Message) {
	replyTo := message.MessageID
	chatID := message.Chat.ID

	lg = lg.With().Int64("chat-id", chatID).Int64("reply-to-id", int64(replyTo)).Logger()

	URLs := extractURLs(message)

	lg.Debug().Strs("urls", URLs).Msg("Found URLs")

	for _, u := range URLs {
		lgu := lg.With().Str("url", u).Logger()

		cmd := YtDlp(ctx, extractor.wd, "--dump-json", "--quiet", u)

		if e := cmd.Run(); e != nil {
			lgu.Warn().Err(e).Stringer("cmd", cmd).Msg("Failed to execute command")
			continue
		}

		bb := cmd.Stdout.(*bytes.Buffer).Bytes()
		decoder := json.NewDecoder(bytes.NewBuffer(bb))
		for decoder.More() {
			info := new(YtDlpInfo)
			if e := decoder.Decode(info); e != nil {
				lgu.Warn().Err(e).Stringer("cmd", cmd).Msg("Unable to decode response from downloader")
				continue
			}
			lgu.Info().Msg("Successfully found media")
			extractor.publishFound(ctx, lgu, message, info)

		}

	}
}

var mediaMarkup = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("best", commandDownload+"b*+ba"),
		tgbotapi.NewInlineKeyboardButtonData("640", commandDownload+"b*[width=640]+ba"),
		tgbotapi.NewInlineKeyboardButtonData("worst", commandDownload+"worst+ba"),
		tgbotapi.NewInlineKeyboardButtonData("multi", commandDownload+"b*[width=640]+ba/worst/best"),
		tgbotapi.NewInlineKeyboardButtonData("List formats", commandFormats),
	),
)

func (extractor *Extractor) publishFound(ctx context.Context, lg zerolog.Logger, message *tgbotapi.Message, info *YtDlpInfo) {
	replyTo := message.MessageID
	chatID := message.Chat.ID

	// maybe info.OriginalURL ? 🤔
	mc := tgbotapi.NewMessage(chatID, `<a href="`+info.WebpageURL+`">🎥 ⇒ `+htmlReplacer.Replace(info.ExtractorKey+": "+info.Title)+`</a>`)
	mc.ReplyToMessageID = replyTo
	mc.ParseMode = "HTML"
	mc.DisableWebPagePreview = true
	mc.ReplyMarkup = mediaMarkup

	mm, err := extractor.botAPI.Send(mc)
	if err != nil {
		lg.Error().Err(err).Msg("Unable to send information about media")
	}

	if chatID == 180727105 {
		// myFormat := "best[ext=mp4][width<=640][filesize<200M]/worst[ext=mp4]"
		// myFormat := "mp4[width=640]/worst[ext=mp4]/ba"
		myFormat := "b*[width=640]+ba/worst/best"
		go extractor.downloadMedia(ctx, lg, myFormat, &mm)
	}
}

// const outputFileFormat = `%(uploader.:64)s/%(title.:96)s • %(id)s.%(format_id)s.%(ext)s`
const outputFileFormat = `%(title.:96)s • %(id)s.%(format_id)s.%(ext)s`

func (extractor *Extractor) downloadMedia(ctx context.Context, lg zerolog.Logger, format string, message *tgbotapi.Message) {
	format = strings.TrimSpace(format)
	URLs := extractURLs(message)

	if len(URLs) != 1 {
		lg.Warn().Strs("urls", URLs).Msgf("Must be only ONE URLs, but found %d", len(URLs))
	}

	if len(URLs) == 0 {
		return
	}

	u := URLs[0] // the first one!

	lgu := lg.With().Str("url", u).Logger()

	cmdPrepare := YtDlp(ctx, extractor.wd, "--dump-json", "--quiet", "-o", outputFileFormat, u)
	if len(format) > 0 {
		cmdPrepare.Args = append(cmdPrepare.Args, "--format", format)
	}

	if e := cmdPrepare.Run(); e != nil {
		lgu.Warn().Err(e).Stringer("cmd", cmdPrepare).Msg("Failed to execute preparation command")
		mc := tgbotapi.NewMessage(message.Chat.ID, "⚠ Unable to get information for "+u+". Format is "+format)
		mc.DisableWebPagePreview = true
		if _, e := extractor.botAPI.Send(mc); e != nil {
			lg.Error().Err(e).Msg("Error while sending message")
		}
		return
	}

	bb := cmdPrepare.Stdout.(*bytes.Buffer).Bytes()
	decoder := json.NewDecoder(bytes.NewBuffer(bb))
	for decoder.More() {
		info := new(YtDlpInfo)
		if e := decoder.Decode(info); e != nil {
			lgu.Warn().Err(e).Msg("Unable to decode response from downloader")
			continue
		}

		lgd := lg.With().Str("url", info.WebpageURL).Logger()

		mc := tgbotapi.NewMessage(message.Chat.ID, "")
		mc.ReplyToMessageID = message.MessageID
		mc.DisableWebPagePreview = true
		mc.ParseMode = "HTML"

		cmdDownload := YtDlp(ctx, extractor.wd, "-o", outputFileFormat, "--progress", info.WebpageURL)
		if len(format) > 0 {
			cmdDownload.Args = append(cmdDownload.Args, "--format", format)
		}

		ctxDownload, cancelDownload := context.WithCancel(ctx)
		defer cancelDownload()

		go extractor.trackStatus(ctxDownload, lgd, cmdDownload.Stdout.(*bytes.Buffer), message.Chat.ID)
		err := cmdDownload.Run()
		cancelDownload()

		if err != nil {
			lgd.Warn().Err(err).Stringer("cmdDownload", cmdDownload).Msg("Failed to execute download command")
			fmt.Fprintln(os.Stderr, cmdDownload.Stderr.(*bytes.Buffer).String())
			mc.Text = `✖✖ Unable to download <b>` + htmlReplacer.Replace(info.Filename) + "</b>"
		} else {
			hrSize := hrSize(filepath.Join(extractor.wd, info.Filename))
			lgd.Info().Stringer("cmdDownload", cmdDownload).Msg("Successfully downloaded")
			tgtURL := extractor.urlPrefix + url.PathEscape(info.Filename)
			linksURL := extractor.linksPrefix + "" + url.PathEscape(info.Filename)
			mc.Text = `<a href="` + linksURL + `">🎯 ⇒ Downloaded <b>` + htmlReplacer.Replace(info.Filename) + "</b></a> " + hrSize
			mu := tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(
					tgbotapi.NewInlineKeyboardButtonURL(
						"🎬 "+info.Ext+": "+info.Resolution,
						tgtURL),
				))
			mc.ReplyMarkup = &mu
		}

		if _, e := extractor.botAPI.Send(mc); e != nil {
			lgd.Warn().Err(e).Msg("Unable to send result message")
		}

	}
}

func hrSize(filePath string) string {
	if st, e := os.Stat(filePath); e == nil {
		return FileSizeHumanReadable(st.Size())
	}
	return ""
}

const tickerInterval = 5 * time.Second

func (extractor *Extractor) trackStatus(ctx context.Context, lg zerolog.Logger, buffer *bytes.Buffer, chatID int64) {
	ticker := time.NewTicker(tickerInterval)
	defer ticker.Stop()

	text := buffer.String()
	if text == "" {
		text = "Starting download..."
	}
	mc := tgbotapi.NewMessage(chatID, `<pre><code>`+text+`</code></pre>`)
	mc.ParseMode = "HTML"

	statusMessage, err := extractor.botAPI.Send(mc)
	if err != nil {
		lg.Error().Err(err).Msg("Failed to send status message")
	}

	defer func() {
		dmc := tgbotapi.NewDeleteMessage(chatID, statusMessage.MessageID)
		extractor.botAPI.Send(dmc)
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			outputText := string(compactWithCR(buffer.Bytes()))
			if len(outputText) == 0 || text == outputText {
				continue
			}
			text = outputText
			em := tgbotapi.NewEditMessageText(chatID, statusMessage.MessageID, `<pre><code>`+text+`</code></pre>`)
			em.ParseMode = "HTML"
			_, err = extractor.botAPI.Send(em)
			if err != nil {
				lg.Warn().Err(err).Msg("Failed to update status message")
			}
		}
	}
}

// compact text with CR support (like any terminal this does)
func compactWithCR(src []byte) []byte {
	result := []byte{}

	if len(src) == 0 {
		return result
	}

	from := 0
	for i := range src {
		if src[i] == '\n' {
			result = append(result, src[from:i+1]...)
			from = i + 1
			continue
		}
		if src[i] == '\r' {
			from = i + 1
			continue
		}
	}

	return append(result, src[from:]...)
}

func (extractor *Extractor) showFormats(ctx context.Context, lg zerolog.Logger, message *tgbotapi.Message) {
	URLs := extractURLs(message)

	if len(URLs) != 1 {
		lg.Warn().Strs("urls", URLs).Msgf("Must be only ONE URLs, but found %d", len(URLs))
	}

	if len(URLs) == 0 {
		return
	}

	u := URLs[0] // the first one!

	lgu := lg.With().Str("url", u).Logger()

	cmdFormat := YtDlp(ctx, extractor.wd, "--list-formats", u)

	if e := cmdFormat.Run(); e != nil {
		lgu.Warn().Err(e).Stringer("cmd", cmdFormat).Msg("Failed to execute preparation command")
		mc := tgbotapi.NewMessage(message.Chat.ID, "⚠ Unable to get format information for "+u)
		mc.DisableWebPagePreview = true
		if _, e := extractor.botAPI.Send(mc); e != nil {
			lg.Error().Err(e).Msg("Error while sending message")
		}
		return
	}

	out := cmdFormat.Stdout.(*bytes.Buffer).String()

	mc := tgbotapi.NewMessage(message.Chat.ID, "<pre><code>"+htmlReplacer.Replace(out)+"</code></pre>")
	mc.DisableWebPagePreview = true
	mc.ParseMode = "HTML"

	if _, e := extractor.botAPI.Send(mc); e != nil {
		lgu.Warn().Err(e).Msg("Unable to send result message")
	}
}

var metrics = map[string]int64{
	"Kb": 1024,
	"Mb": 1024 * 1024,
	"Gb": 1024 * 1024 * 1024,
	"Tb": 1024 * 1024 * 1024 * 1024,
}

func FileSizeHumanReadable(size int64) string {
	for _, m := range []string{"Tb", "Gb", "Mb", "Kb"} {
		if size >= metrics[m] {
			return fmt.Sprintf("%.1f%s", float64(float64(size)/float64(metrics[m])), m)
		}
	}
	return fmt.Sprintf("%db", size)
}
