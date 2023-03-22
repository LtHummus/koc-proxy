package proxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"

	"github.com/lthummus/koc-proxy/auth"
	"github.com/lthummus/koc-proxy/types"

	"github.com/rs/zerolog/log"
)

// NewAuthRequestInterceptor takes an already built reverse proxy and attaches a handler to sniff out requests to
// POST /api/auth and intercept them in order to do our authentication magic. All other requests are passed through
// unchanged
func NewAuthRequestInterceptor(prx *httputil.ReverseProxy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug().Str("request_uri", r.RequestURI).Str("method", r.Method).Msg("request")
		if r.RequestURI != "/api/auth" || r.Method != http.MethodPost {
			// not a request we care about? pass it on and we're done
			prx.ServeHTTP(w, r)
			return
		}

		payloadBytes, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).Msg("could not read payload")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		r.Body.Close()

		clientIP, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			log.Error().Err(err).Msg("could not parse remote IP")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var authPayload types.AuthPayload
		err = json.Unmarshal(payloadBytes, &authPayload)
		if err != nil {
			log.Error().Err(err).Msg("could not decode payload")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		valid := auth.CheckAuth(r.Context(), authPayload.Credentials.Username, authPayload.Credentials.Secret)
		if !valid {
			// authentication secret invalid, log the failure, then send back the same response the backend server would have sent
			log.Warn().Str("source_ip", clientIP).Str("username", authPayload.Credentials.Username).Msg("invalid password")
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Length", fmt.Sprintf("%d", types.IncorrectPasswordResponseLength))

			// i'm not sure why this is HTTP 422 as a response instead of 403, but whatever, i'm just following what the
			// server does
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write(types.IncorrectPasswordResponse)
			return
		} else {
			// correct response, take the request payload, patch the proper secret in to it (i.e. the one the server actually
			// expects) and pass it on through, then pass the response back to the client
			log.Info().Str("source_ip", clientIP).Str("username", authPayload.Credentials.Username).Msg("successful auth")
			patched := authPayload
			patched.Credentials.Secret = auth.ServerSecret()
			patchedPayload, err := json.Marshal(patched)
			if err != nil {
				log.Error().Err(err).Msg("could not reencode payload")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			patchedRequest, err := http.NewRequest(r.Method, r.RequestURI, bytes.NewReader(patchedPayload))
			if err != nil {
				log.Error().Err(err).Msg("could not build patched request")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			patchedRequest.Header.Set("Content-Length", fmt.Sprintf("%d", len(patchedPayload)))
			prx.ServeHTTP(w, patchedRequest)
		}

	}
}
