package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/lthummus/koc-proxy/authdb"
	"github.com/lthummus/koc-proxy/proxy"
	"github.com/lthummus/koc-proxy/util"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var proxyCommand = &cobra.Command{
	Use:   "proxy",
	Short: "Run the KOC Standalone Server Auth Proxy",
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().Msg("Starting in proxy mode")

		err := authdb.Connect(cmd.Context())
		util.FatalIfError(err, "could not connect to auth db")

		prx := proxy.MakeProxy(fmt.Sprintf("http://%s", viper.GetString("proxy.backend.upstream")))
		http.HandleFunc("/", proxy.NewAuthRequestInterceptor(prx))

		port := viper.GetInt("proxy.port")
		if port == 0 {
			port = 23600
		}

		log.Info().Int("port", port).Msg("starting server")

		listenAddress := fmt.Sprintf(":%d", port)

		srv := http.Server{
			Addr: listenAddress,
		}

		allCleaned := make(chan struct{})
		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)
			<-sigChan
			log.Warn().Msg("got interrupted; shutting down")

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := srv.Shutdown(ctx); err != nil {
				log.Error().Err(err).Msg("error on shutdown")
			}

			close(allCleaned)
		}()

		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Err(err).Msg("error starting server")
		}

		<-allCleaned

		err = authdb.Disconnect(cmd.Context())
		util.FatalIfError(err, "could not disconnect from auth db")

		log.Info().Msg("shutting down")
	},
}
