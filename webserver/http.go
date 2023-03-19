package webserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

func RunHTTPServer() {
	webPort := viper.GetInt("web.port")
	if webPort == 0 {
		webPort = 8080
	}

	log.Info().Int("port", webPort).Msg("running in web server mode")

	srv := buildServer(fmt.Sprintf(":%d", webPort))

	srv.Handler = buildMux()

	allCleaned := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan

		log.Warn().Msg("starting shutdown")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("err on shutdown")
		}

		close(allCleaned)
	}()

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Error().Err(err).Msg("error starting server")
	}

	<-allCleaned

	log.Info().Msg("shutting down")
}
