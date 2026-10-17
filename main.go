// Command w42-eu-web serves the w42.eu landing page: a list of mj41's projects, with links to
// mj41.cz and GitHub. It also serves mcbot.w42.eu, the Minecraft robots' page with its
// screenshots (/mcbot/*.png, /mcbot/*.webp). The pages are embedded and chosen by the request's
// host.
package main

import (
	"embed"
	"flag"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

//go:embed index.html
var indexHTML []byte

//go:embed mcbot.html
var mcbotHTML []byte

// images are mcbot.html's screenshots, served at /mcbot/<name>, and index.html's pictures,
// served at /img/<name>.
//
//go:embed mcbot/*.png mcbot/*.webp img/*.webp
var images embed.FS

// imageTypes are the images' content types by extension.
var imageTypes = map[string]string{".png": "image/png", ".webp": "image/webp"}

// pages maps a host to its page; any other host gets the w42.eu page.
var pages = map[string][]byte{
	"mcbot.w42.eu": mcbotHTML,
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

func writePage(w http.ResponseWriter, p []byte) {
	h := w.Header()
	h.Set("Content-Type", "text/html; charset=utf-8")
	h.Set("Cache-Control", "public, max-age=300")
	h.Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; img-src 'self' data:; base-uri 'none'; form-action 'none'; frame-ancestors 'none'")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
	w.Write(p)
}

func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		writePage(w, page(r.Host))
	})
	// one file by name: no listing, and {name} cannot hold a "/"
	for _, dir := range []string{"mcbot", "img"} {
		mux.HandleFunc("GET /"+dir+"/{name}", func(w http.ResponseWriter, r *http.Request) {
			name := r.PathValue("name")
			ct := imageTypes[path.Ext(name)]
			b, err := images.ReadFile(dir + "/" + name)
			if ct == "" || err != nil {
				http.NotFound(w, r)
				return
			}
			h := w.Header()
			h.Set("Content-Type", ct)
			h.Set("Cache-Control", "public, max-age=86400")
			h.Set("X-Content-Type-Options", "nosniff")
			w.Write(b)
		})
	}
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok\n")) })
	return mux
}

func main() {
	listen := flag.String("listen", ":8080", "HTTP listen address")
	flag.Parse()
	log := slog.New(slog.NewTextHandler(os.Stderr, nil))

	srv := &http.Server{
		Addr:              *listen,
		Handler:           handler(),
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
