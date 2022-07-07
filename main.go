package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"almaz.uno/dev/almaz-video-bot/pkg/loghook"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	cfgLevel     = os.Getenv("LEVEL")
	cfgStackPath = os.Getenv("STACK_PATH")
	cfgCaller    = os.Getenv("CALLER") == "true"
	cfgToken     = os.Getenv("TOKEN")
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

	if cfgToken == "" {
		log.Fatal().Msg("TOKEN environment variable is required")
	}
	doMain(func(ctx context.Context) error {
		return loop(ctx, cfgToken)
	})
}

// doMain starts function runFunc with context. The context will be canceled
// by SIGTERM or SIGINT signal (Ctrl+C for example)
func doMain(runFunc func(ctx context.Context) error) {
	// context should be canceled while Int signal will be caught
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// main processing loop
	retChan := make(chan error, 1)
	go func() {
		err2 := runFunc(ctx)
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
