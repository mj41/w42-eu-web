package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"testing"
)

func TestPage(t *testing.T) {
	for host, want := range map[string][]byte{
		"w42.eu":           indexHTML,
		"localhost:8080":   indexHTML,
		"home.w42.eu":      indexHTML,
		"S.w42.eu:443":     sHTML,
		"s.w42.eu.evil":    indexHTML,
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

func TestImages(t *testing.T) {
	srv := httptest.NewServer(handler())
	defer srv.Close()
	for p, want := range map[string]int{
		"/mcbot/dashboard.png":       200,
		"/mcbot/dashboard-small.png": 200,
		"/mcbot/":                    404,
		"/mcbot/nope.png":            404,
		"/mcbot/../main.go":          404,
		"/mcbot/%2e%2e%2fmain.go":    404,
		"/s/robot-focus.webp":        200,
		"/s/":                        404,
		"/s/nope.webp":               404,
		"/main.go":                   404,
		"/x/main.go":                 404,
		"/s/../go.mod":               404,
	} {
		res, err := http.Get(srv.URL + p)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		if res.StatusCode != want {
			t.Errorf("%s: %d, want %d", p, res.StatusCode, want)
		}
		if want == 200 && res.Header.Get("Content-Type") != imageTypes[path.Ext(p)] {
			t.Errorf("%s: %s", p, res.Header.Get("Content-Type"))
		}
	}
	// every image is named by its page, and every image a page names is there
	for dir, html := range map[string][]byte{"mcbot": mcbotHTML, "s": sHTML} {
		names, _ := images.ReadDir(dir)
		for _, n := range names {
			if !bytes.Contains(html, []byte("/"+dir+"/"+n.Name())) {
				t.Errorf("%s.html does not name %s", dir, n.Name())
			}
		}
	}
	for _, m := range regexp.MustCompile(`"/(s|mcbot)/([^"]+)"`).FindAllSubmatch(append(append([]byte{}, sHTML...), append(mcbotHTML, indexHTML...)...), -1) {
		if _, err := images.ReadFile(string(m[1]) + "/" + string(m[2])); err != nil {
			t.Errorf("a page names a missing image: /%s/%s", m[1], m[2])
		}
	}
}
