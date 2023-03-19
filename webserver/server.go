package webserver

import (
	"net/http"
	"time"
)

func buildServer(addr string) *http.Server {
	return &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  2 * time.Minute,
	}
}

func buildMux() *http.ServeMux {
	mux := &http.ServeMux{}
	mux.HandleFunc("/", RenderStatusPage)
	mux.HandleFunc("/ping", RenderPingPage)

	return mux
}
