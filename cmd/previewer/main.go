package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	internals3 "github.com/spendmail/s3_previewer/internal/s3"
	internalapp "github.com/spendmail/s3_previewer/internal/app"
	internalcache "github.com/spendmail/s3_previewer/internal/cache"
	internalconfig "github.com/spendmail/s3_previewer/internal/config"
	internallogger "github.com/spendmail/s3_previewer/internal/logger"
	internalresizer "github.com/spendmail/s3_previewer/internal/resizer"
	internalserver "github.com/spendmail/s3_previewer/internal/server/http"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "/etc/s3_previewer/previewer.toml", "Path to configuration file")
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config, err := internalconfig.NewConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	logger, err := internallogger.New(config)
	if err != nil {
		log.Fatal(err)
	}

	cache, err := internalcache.New(config, logger)
	if err != nil {
		log.Fatal(err)
	}

	s3Client, err := internals3.New(config, logger)
	if err != nil {
		log.Fatal(err)
	}

	app, err := internalapp.New(logger, internalresizer.New(), cache, s3Client)
	if err != nil {
		log.Fatal(err)
	}

	server := internalserver.New(config, logger, app)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGHUP)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		// Locking until OS signal is sent or context cancel func is called.
		<-ctx.Done()

		// Stopping http server.
		stopHTTPCtx, stopHTTPCancel := context.WithTimeout(context.Background(), time.Second*3)
		defer stopHTTPCancel()
		if err := server.Stop(stopHTTPCtx); err != nil {
			logger.Error(err.Error())
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		logger.Info("starting http server...")

		// Locking over here until server is stopped.
		if err := server.Start(); err != nil {
			logger.Error(err.Error())
			cancel()
		}
	}()

	wg.Wait()
}
