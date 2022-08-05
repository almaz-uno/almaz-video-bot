package main

import (
	"context"
	_ "embed"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"almaz.uno/dev/almaz-video-bot/pkg/extractors/mediadl"
	"almaz.uno/dev/almaz-video-bot/pkg/loghook"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	cfgLevel        = os.Getenv("LEVEL")
	cfgStackPath    = os.Getenv("STACK_PATH")
	cfgCaller       = os.Getenv("CALLER") == "true"
	cfgToken        = os.Getenv("TOKEN")
	cfgMediaDir     = os.Getenv("MEDIA_DIR")
	cfgServerPrefix = os.Getenv("SERVER_PREFIX")
	cfgListenOn     = os.Getenv("LISTEN_ON")
	cfgCertFile     = os.Getenv("CERT_FILE")
	cfgKeyFile      = os.Getenv("KEY_FILE")
	cfgStaticPrefix = "/media/"
	cfgLinksPrefix  = "/links/"
)

var (
	//go:embed list.html
	tmplList string

	//go:embed links.html
	tmplLinks string
)

func main() {
	if level, e := zerolog.ParseLevel(cfgLevel); e == nil {
		zerolog.SetGlobalLevel(level)
	}

	log.Logger = log.Hook(&loghook.GoroutineStack{
		GIDName:   "gid",
		StackFile: "stack-path",
		StackPath: cfgStackPath,
	})

	if cfgCaller {
		log.Logger = log.With().Caller().Logger()
	}

	if cfgMediaDir == "" {
		cfgMediaDir = "."
	} else {
		os.MkdirAll(cfgMediaDir, 0o755)
	}

	if cfgToken == "" {
		log.Fatal().Msg("TOKEN environment variable is required")
	}

	ec := echo.New()
	ec.Static(cfgStaticPrefix, cfgMediaDir)
	ec.GET("/list", list)
	ec.GET("/links/*", links)

	doMain(func(ctx context.Context, cancel context.CancelFunc) error {
		go func() {
			defer cancel()

			forceTLS := strings.HasPrefix(cfgServerPrefix, "https://")

			var e error
			if forceTLS {
				e = ec.StartTLS(cfgListenOn, cfgCertFile, cfgKeyFile)
			} else {
				e = ec.Start(cfgListenOn)
			}
			if e != nil && !errors.Is(e, http.ErrServerClosed) {
				log.Error().Err(e).Msg("Unable to start echo server")
			}
		}()

		loop(ctx, cfgToken)

		closeCtx, closeCancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer closeCancel()

		return ec.Shutdown(closeCtx)
	})
}

// doMain starts function runFunc with context. The context will be canceled
// by SIGTERM or SIGINT signal (Ctrl+C for example)
func doMain(runFunc func(ctx context.Context, cancel context.CancelFunc) error) {
	// context should be canceled while Int signal will be caught
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// main processing loop
	retChan := make(chan error, 1)
	go func() {
		err2 := runFunc(ctx, cancel)
		if err2 != nil {
			retChan <- err2
		}
		close(retChan)
	}()

	// Waiting signals from OS
	go func() {
		quit := make(chan os.Signal, 10)
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		log.Warn().Msgf("Signal '%s' was caught. Exiting", <-quit)
		cancel()
	}()

	// Listening for the main loop response
	for e := range retChan {
		log.Info().Err(e).Msg("Exiting.")
	}
}

func list(c echo.Context) error {
	files := []fileInfo{}
	var totalSize int64
	root, _ := filepath.Abs(cfgMediaDir)

	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsRegular() {
			p := path
			if strings.HasPrefix(p, root) {
				p = p[len(root)+1:]
			}

			fURL, err := url.Parse(cfgServerPrefix + cfgStaticPrefix + p)
			if err != nil {
				panic(err)
			}

			lURL, err := url.Parse(cfgServerPrefix + cfgLinksPrefix + p)
			if err != nil {
				panic(err)
			}

			f := fileInfo{
				d:        d,
				Path:     p,
				URL:      fURL.String(),
				LinksURL: lURL.String(),
			}
			totalSize += f.Size()

			files = append(files, f)
		}
		return nil
	})

	sort.Slice(files, func(i, j int) bool {
		return files[i].ATime().After(files[j].ATime())
	})

	context := map[string]any{
		"files":        files,
		"TotalSizeStr": mediadl.FileSizeHumanReadable(totalSize),
	}

	return template.Must(template.New("list").Parse(tmplList)).Execute(c.Response().Writer, context)
}

func links(c echo.Context) error {
	u := c.Request().URL

	p := u.Path
	if strings.HasPrefix(p, cfgLinksPrefix) {
		p = p[len(cfgLinksPrefix):]
	}

	downloadURL, err := url.Parse(cfgServerPrefix + cfgStaticPrefix + p)
	if err != nil {
		panic(err)
	}

	context := map[string]any{
		"Title": p,
		"URL":   downloadURL,
	}

	return template.Must(template.New("links").Parse(tmplLinks)).Execute(c.Response().Writer, context)
}
