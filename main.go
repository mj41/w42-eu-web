// Command w42-eu-web serves the w42.eu landing page: a list of the projects that
// run under w42.eu, with links to mj41.cz and GitHub. It also serves home.w42.eu,
// a page pointing to the home-w42-eu repositories, and s.w42.eu, the Stackchan
// project page. The pages are embedded and chosen by the request's host.
package main

import (
	_ "embed"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

//go:embed index.html
var indexHTML []byte

//go:embed home.html
var homeHTML []byte

//go:embed s.html
var sHTML []byte

// pages maps a host to its page; any other host gets the w42.eu page.
var pages = map[string][]byte{
	"home.w42.eu": homeHTML,
	"s.w42.eu":    sHTML,
}

func page(host string) []byte {
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	if p, ok := pages[strings.ToLower(host)]; ok {
		return p
	}
	return indexHTML
}

func main() {
	listen := flag.String("listen", ":8080", "HTTP listen address")
	flag.Parse()
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Cache-Control", "public, max-age=300")
		h.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src 'self' data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Write(page(r.Host))
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })

	srv := &http.Server{
		Addr:              *listen,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Info("w42-eu-web listening", "listen", *listen)
	if err := srv.ListenAndServe(); err != nil {
		log.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
