package proxy

import (
	"net/http/httputil"
	"net/url"

	"github.com/rs/zerolog/log"
)

// MakeProxy generates a reverse proxy implementation given an upstream url to contact
func MakeProxy(upstream string) *httputil.ReverseProxy {
	target, err := url.Parse(upstream)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse target url")
	}
	log.Info().Str("upstream", target.String()).Msg("starting proxy")
	return httputil.NewSingleHostReverseProxy(target)
}
