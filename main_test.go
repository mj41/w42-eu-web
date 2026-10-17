package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"path"
	"regexp"
	"strings"
	"testing"
)

func TestPage(t *testing.T) {
	for host, want := range map[string][]byte{
		"w42.eu":           indexHTML,
		"localhost:8080":   indexHTML,
		"home.w42.eu":      indexHTML,
		"S.w42.eu:443":     indexHTML,
		"s.i41.eu.evil":    indexHTML,
		"s.w42.eu":         indexHTML,
		"s.i41.eu":         indexHTML, // i41eu-web's now
		"i41.eu":           indexHTML,
		"m.i41.eu":         indexHTML,
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
		"/mcbot/dashboard.png":        200,
		"/mcbot/dashboard-small.webp": 200,
		"/img/i41-robot.webp":         200,
		"/img/infinite-pm.webp":       200,
		"/img/":                       404,
		"/img/nope.webp":              404,
		"/img/%2e%2e%2fmain.go":       404,
		"/mcbot/":                     404,
		"/mcbot/nope.png":             404,
		"/mcbot/../main.go":           404,
		"/mcbot/%2e%2e%2fmain.go":     404,
		"/s/robot3d-focus.png":        404, // i41eu-web's now
		"/main.go":                    404,
		"/x/main.go":                  404,
		"/s/../go.mod":                404,
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
	// every image is named by mcbot.html, and every image a page names is there
	names, _ := images.ReadDir("mcbot")
	for _, n := range names {
		if !bytes.Contains(mcbotHTML, []byte("/mcbot/"+n.Name())) {
			t.Errorf("mcbot.html does not name %s", n.Name())
		}
	}
	// every image in img/ is named by index.html
	names, _ = images.ReadDir("img")
	for _, n := range names {
		if !bytes.Contains(indexHTML, []byte("/img/"+n.Name())) {
			t.Errorf("index.html does not name %s", n.Name())
		}
	}
	for name, page := range allPages() {
		for _, m := range regexp.MustCompile(`"/([a-z]+)/([^"]+\.(png|webp))"`).FindAllSubmatch(page, -1) {
			if _, err := images.ReadFile(string(m[1]) + "/" + string(m[2])); err != nil {
				t.Errorf("%s names a missing image: /%s/%s", name, m[1], m[2])
			}
		}
	}
}

// Links to other sites open in a new tab, so the visitor keeps this page; links to our own
// sites (w42.eu, i41.eu, q41.eu and their subdomains, mj41.cz) stay in the tab.
func TestExternalLinksOpenANewTab(t *testing.T) {
	link := regexp.MustCompile(`<a\b[^>]*\bhref="https?://([^/"]+)[^"]*"[^>]*>`)
	ours := regexp.MustCompile(`(^|\.)(w42\.eu|i41\.eu|q41\.eu|mj41\.cz)$`)
	for name, page := range allPages() {
		for _, m := range link.FindAllSubmatch(page, -1) {
			tag, host := string(m[0]), string(m[1])
			newTab := bytes.Contains(m[0], []byte(`target="_blank"`)) && bytes.Contains(m[0], []byte(`rel="noopener"`))
			if ours.MatchString(host) == newTab {
				t.Errorf("%s: %s: our sites stay in the tab, others open a new one (with rel=noopener)", name, tag)
			}
			// another site says so after the link: a chip (data-ext), unless its text says it already
			chip := bytes.Contains(m[0], []byte(`data-ext="`))
			if ours.MatchString(host) && chip {
				t.Errorf("%s: %s: a chip on a link to our own site", name, tag)
			}
		}
	}
}

// Every link to another site is marked after it: a chip with the site ("GitHub", "m5stack.com"),
// unless the link's own text names it ("github.com/mj41", "infinite.pm").
func TestExternalLinksMarked(t *testing.T) {
	link := regexp.MustCompile(`(?s)<a\b([^>]*\bhref="https?://([^/"]+)[^"]*"[^>]*)>(.*?)</a>`)
	ours := regexp.MustCompile(`(^|\.)(w42\.eu|i41\.eu|q41\.eu|mj41\.cz)$`)
	for name, page := range allPages() {
		for _, m := range link.FindAllSubmatch(page, -1) {
			attrs, host, text := string(m[1]), strings.ToLower(string(m[2])), strings.ToLower(string(m[3]))
			if ours.MatchString(host) || strings.Contains(attrs, `data-ext="`) {
				continue
			}
			site := strings.TrimPrefix(strings.TrimPrefix(host, "www."), "shop.")
			if strings.HasSuffix(host, "github.com") {
				site = "github"
			}
			if !strings.Contains(text, site) {
				t.Errorf("%s: %q to %s has no chip and does not say where it goes", name, text, host)
			}
		}
	}
}

// allPages are every page served, by name.
func allPages() map[string][]byte {
	return map[string][]byte{"index.html": indexHTML, "mcbot.html": mcbotHTML}
}

// No app pages here any more: they are i41eu-web's (s.i41.eu/<app>).
func TestNoAppPages(t *testing.T) {
	srv := httptest.NewServer(handler())
	defer srv.Close()
	for _, host := range []string{"w42.eu", "s.i41.eu"} {
		for _, p := range []string{"/pet", "/focus", "/raw", "/dev"} {
			req, _ := http.NewRequest("GET", srv.URL+p, nil)
			req.Host = host
			res, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			res.Body.Close()
			if res.StatusCode != 404 {
				t.Errorf("%s%s: %d, want 404", host, p, res.StatusCode)
			}
		}
	}
}
