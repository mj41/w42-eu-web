package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPage(t *testing.T) {
	for host, want := range map[string][]byte{
		"w42.eu":           indexHTML,
		"localhost:8080":   indexHTML,
		"home.w42.eu":      homeHTML,
		"HOME.w42.eu:443":  homeHTML,
		"home.w42.eu.evil": indexHTML,
		"s.w42.eu":         sHTML,
		"mcbot.w42.eu":     mcbotHTML,
		"mon.mcbot.w42.eu": indexHTML,
		"sm.w42.eu":        indexHTML,
	} {
		if !bytes.Equal(page(host), want) {
			t.Errorf("page(%q): wrong page", host)
		}
	}
}

func TestMcbotImages(t *testing.T) {
	srv := httptest.NewServer(handler())
	defer srv.Close()
	for path, want := range map[string]int{
		"/mcbot/dashboard.png":       200,
		"/mcbot/dashboard-small.png": 200,
		"/mcbot/":                    404,
		"/mcbot/nope.png":            404,
		"/mcbot/../main.go":          404,
		"/mcbot/%2e%2e%2fmain.go":    404,
	} {
		res, err := http.Get(srv.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != want {
			t.Errorf("%s: %d, want %d", path, res.StatusCode, want)
		}
		if want == 200 && res.Header.Get("Content-Type") != "image/png" {
			t.Errorf("%s: %s", path, res.Header.Get("Content-Type"))
		}
	}
	// every image the page names is served
	for _, name := range []string{"dashboard.png", "dashboard-small.png"} {
		if !bytes.Contains(mcbotHTML, []byte("/mcbot/"+name)) {
			t.Errorf("mcbot.html does not name %s", name)
		}
	}
}
