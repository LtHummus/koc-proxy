package webserver

import (
	"bytes"
	"context"
	_ "embed"
	"html/template"
	"net/http"
	"sync"
	"time"

	"github.com/lthummus/koc-proxy/vbackend"
	"github.com/lthummus/koc-proxy/vredis"

	"github.com/rs/zerolog/log"
)

const cacheTimeout = 1 * time.Minute

//go:embed templates/status.gohtml
var templateText string

var statusTemplate *template.Template

func init() {
	statusTemplate = template.Must(template.New("").Parse(templateText))
}

var (
	cachedPage []byte
	cacheTime  = time.Now().AddDate(-10, 0, 0)

	cacheLock = &sync.Mutex{}
)

func buildPageParams(ctx context.Context) (map[string]any, error) {
	ret := map[string]any{}

	up, err := vbackend.IsAlive(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not check to see liveness")
		return ret, err
	}

	ret["up"] = up

	if !up {
		return ret, nil
	}

	players, err := vredis.GetConnectedCount(ctx)
	if err != nil {
		log.Error().Err(err).Msg("could not check to count players")
		return ret, err
	}

	ret["players"] = players
	return ret, nil
}

func renderPage(ctx context.Context) ([]byte, error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if time.Since(cacheTime) > cacheTimeout || cachedPage == nil {
		log.Trace().Msg("rendering page since cache expired")

		data, err := buildPageParams(ctx)
		if err != nil {
			return nil, err
		}

		var output bytes.Buffer
		if err := statusTemplate.Execute(&output, data); err != nil {
			return nil, err
		}
		cachedPage = output.Bytes()
		cacheTime = time.Now()
	}

	return cachedPage, nil
}

func RenderStatusPage(w http.ResponseWriter, r *http.Request) {
	payload, err := renderPage(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("could not render page")
		http.Error(w, "could not render page", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")

	w.WriteHeader(http.StatusOK)

	w.Write(payload)
}

func RenderPingPage(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello world!"))
}
