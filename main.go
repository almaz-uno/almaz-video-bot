package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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

	// Слушаем сигналы завершения от ОС. В разных ОС они работают по-разному.
	// При поступлении завершающего сигнала ОС по контексту исполнения
	// передаёт сигнал останова для изящного завершения
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
