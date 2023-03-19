package webserver

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/lthummus/koc-proxy/util"

	"golang.org/x/crypto/acme"
	"golang.org/x/crypto/acme/autocert"
	"golang.org/x/sync/errgroup"

	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type tlsConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	Port     int    `mapstructure:"port"`
	Domain   string `mapstructure:"domain"`
	CacheDir string `mapstructure:"cache"`
	Test     bool   `mapstructure:"test"`
}

func RunHTTPSServer() {
	var cfg tlsConfig
	err := viper.UnmarshalKey("web.tls", &cfg)
	util.FatalIfError(err, "could not read TLS config")
	log.
		Info().
		Str("domain", cfg.Domain).
		Int("port", cfg.Port).
		Str("cache_dir", cfg.CacheDir).
		Bool("test_mode", cfg.Test).
		Msg("running with TLS enabled")

	certManager := autocert.Manager{
		Prompt: autocert.AcceptTOS,
		Cache:  autocert.DirCache(cfg.CacheDir),
		HostPolicy: func(ctx context.Context, host string) error {
			if host == cfg.Domain {
				return nil
			}

			log.Warn().Str("host", host).Msg("invalid host requested")
			return fmt.Errorf("invalid cert requested: %s", host)
		},
	}

	if cfg.Test {
		log.Warn().Msg("enabling test mode for letsencrypt")
		// if we're testing cert stuff, use the staging letsencrypt environment so we don't
		// blow away our request limits
		certManager.Client = &acme.Client{
			DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
		}
	}

	tlsServer := buildServer(fmt.Sprintf(":%d", cfg.Port))
	tlsServer.TLSConfig = &tls.Config{
		GetCertificate: certManager.GetCertificate,
	}
	tlsServer.Handler = buildMux()

	plaintextServer := buildServer(fmt.Sprintf(":%d", viper.GetInt("web.port")))
	plaintextServer.Handler = certManager.HTTPHandler(nil) // TODO: redirect

	plainCleaned := make(chan struct{})
	tlsCleaned := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt)
		<-sigChan

		log.Warn().Msg("starting shutdown")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		errGroup, gctx := errgroup.WithContext(ctx)
		errGroup.Go(func() error {
			if perr := plaintextServer.Shutdown(gctx); perr != nil {
				log.Error().Err(perr).Msg("error shutting down plaintext server")
				return perr
			}

			log.Info().Msg("plaintext server shutdown")

			return nil
		})

		errGroup.Go(func() error {
			if terr := tlsServer.Shutdown(gctx); terr != nil {
				log.Error().Err(terr).Msg("error shutting down tls server")
				return terr
			}

			log.Info().Msg("tls server shutdown")

			return nil
		})

		err = errGroup.Wait()
		if err != nil {
			log.Warn().Msg("error shutting down servers")
		}

		close(tlsCleaned)
		close(plainCleaned)

	}()

	go func() {
		if terr := tlsServer.ListenAndServeTLS("", ""); terr != http.ErrServerClosed {
			log.Error().Err(terr).Msg("error starting tls server")
		}
	}()

	go func() {
		if perr := plaintextServer.ListenAndServe(); perr != http.ErrServerClosed {
			log.Error().Err(perr).Msg("error starting plaintext server")
		}
	}()

	<-tlsCleaned
	<-plainCleaned

	log.Info().Msg("shutting down")
}
