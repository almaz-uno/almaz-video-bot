package main

import (
	"context"
	"errors"
	"html/template"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
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

const tmplList = `
<html>
	<head>
		<title>File list</title>
	</head>

	<body>
		<table>
			<tr>
				<th>File</th>
				<th>Size</th>
				<th>MTime</th>
				<th>CTime</th>
				<th>ATime</th>
			</tr>
			{{range .files}}
			<tr>
				<td><a href="{{.URL}}">{{.Name}}</a></td>
				<td>{{.SizeStr}}</td>
				<td>{{.MTimeAgo}}</td>
				<td>{{.CTimeAgo}}</td>
				<td>{{.ATimeAgo}}</td>
			</tr>
			{{else}}
			there no files yet
			{{end}}
		</table>
	</body>
</html>
`

const timeFormat = "2006-01-02 15:04:05Z07:00"

type fileInfo struct {
	d   fs.DirEntry
	URL string
}

func (fi fileInfo) Name() string {
	return fi.d.Name()
}

func (fi fileInfo) SizeStr() string {
	return mediadl.FileSizeHumanReadable(fi.Size())
}

func (fi fileInfo) ATimeAgo() time.Duration {
	return time.Duration(time.Since(fi.ATime()).Seconds()) * time.Second
}

func (fi fileInfo) CTimeAgo() time.Duration {
	return time.Duration(time.Since(fi.CTime()).Seconds()) * time.Second
}

func (fi fileInfo) MTimeAgo() time.Duration {
	return time.Duration(time.Since(fi.MTime()).Seconds()) * time.Second
}

func (fi fileInfo) ATimeStr() string {
	return fi.ATime().Format(timeFormat)
}

func (fi fileInfo) CTimeStr() string {
	return fi.CTime().Format(timeFormat)
}

func (fi fileInfo) MTimeStr() string {
	return fi.MTime().Format(timeFormat)
}

func (fi fileInfo) MTime() time.Time {
	i, err := fi.d.Info()
	if err != nil {
		log.Warn().Err(err).Str("file", fi.d.Name()).
			Msg("Unable to get file info")
		return time.Time{}
	}
	return i.ModTime()
}

func (fi fileInfo) ATime() time.Time {
	i, err := fi.d.Info()
	if err != nil {
		log.Warn().Err(err).Str("file", fi.d.Name()).
			Msg("Unable to get file info")
		return time.Time{}
	}
	stat := i.Sys().(*syscall.Stat_t)
	return time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))
}

func (fi fileInfo) CTime() time.Time {
	i, err := fi.d.Info()
	if err != nil {
		log.Warn().Err(err).Str("file", fi.d.Name()).
			Msg("Unable to get file info")
		return time.Time{}
	}
	stat := i.Sys().(*syscall.Stat_t)
	return time.Unix(int64(stat.Ctim.Sec), int64(stat.Ctim.Nsec))
}

func (fi fileInfo) Size() int64 {
	i, err := fi.d.Info()
	if err != nil {
		log.Warn().Err(err).Str("file", fi.d.Name()).
			Msg("Unable to get file info")
		return 0
	}
	return i.Size()
}

func list(c echo.Context) error {
	files := []fileInfo{}
	filepath.WalkDir(cfgMediaDir, func(path string, d fs.DirEntry, err error) error {
		if d.Type().IsRegular() {
			files = append(files, fileInfo{
				d:   d,
				URL: cfgServerPrefix + cfgStaticPrefix + url.PathEscape(d.Name()),
			})
		}
		return nil
	})

	context := map[string]any{
		"files": files,
	}

	return template.Must(template.New("list").Parse(tmplList)).Execute(c.Response().Writer, context)
}
