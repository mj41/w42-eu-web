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
		"w42.eu":           indexPage,
		"localhost:8080":   indexPage,
		"home.w42.eu":      indexPage,
		"S.w42.eu:443":     indexPage,
		"s.i41.eu.evil":    indexPage,
		"s.w42.eu":         indexPage,
		"s.i41.eu":         indexPage, // i41eu-web's now
		"i41.eu":           indexPage,
		"m.i41.eu":         indexPage,
		"mcbot.w42.eu":     mcbotPage,
		"mon.mcbot.w42.eu": indexPage,
		"sm.w42.eu":        indexPage,
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
		"/mcbot/acam-1-dusk-plains.webp":         200,
		"/mcbot/acam-1-dusk-plains-small.webp":   200,
		"/mcbot/acam-2-harvest-field.webp":       200,
		"/mcbot/acam-2-harvest-field-small.webp": 200,
		"/mcbot/acam-3-harvest-wheat.webp":       200,
		"/mcbot/acam-3-harvest-wheat-small.webp": 200,
		"/mcbot/dashboard.png":                   404, // the dashboard is private; no screenshot
		"/mcbot/dashboard-small.webp":            404,
		"/img/i41-robot.webp":                    200,
		"/img/infinite-pm.webp":                  200,
		"/img/":                                  404,
		"/img/nope.webp":                         404,
		"/img/%2e%2e%2fmain.go":                  404,
		"/mcbot/":                                404,
		"/mcbot/nope.png":                        404,
		"/mcbot/../main.go":                      404,
		"/mcbot/%2e%2e%2fmain.go":                404,
		"/s/robot3d-focus.png":                   404, // i41eu-web's now
		"/main.go":                               404,
		"/x/main.go":                             404,
		"/s/../go.mod":                           404,
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
	// the page shows three action-camera views, small, each linking to the full size
	for _, n := range []string{"acam-1-dusk-plains", "acam-2-harvest-field", "acam-3-harvest-wheat"} {
		for _, f := range []string{n + "-small.webp", n + ".webp"} {
			if !bytes.Contains(mcbotHTML, []byte("/mcbot/"+f)) {
				t.Errorf("mcbot.html does not name %s", f)
			}
		}
	}
	if bytes.Contains(mcbotHTML, []byte("/mcbot/dashboard")) || bytes.Contains(indexHTML, []byte("/mcbot/dashboard")) {
		t.Error("a page still names the dashboard screenshot")
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

// Pages name their images by content hash (/img/NAME.HASH.ext), so a changed picture gets a new
// URL and can be cached for a year; the plain name still works, briefly cached.
func TestHashedImageURLs(t *testing.T) {
	srv := httptest.NewServer(handler())
	defer srv.Close()
	get := func(host, p string) (*http.Response, []byte) {
		req, _ := http.NewRequest("GET", srv.URL+p, nil)
		req.Host = host
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		var b bytes.Buffer
		b.ReadFrom(res.Body)
		return res, b.Bytes()
	}
	ref := regexp.MustCompile(`(?:src|href)="(/(?:img|mcbot)/[^"]+\.(?:png|webp))"`)
	hashed := regexp.MustCompile(`^/(img|mcbot)/([^/]+)\.[0-9a-f]{8}(\.(?:png|webp))$`)
	for _, host := range []string{"w42.eu", "mcbot.w42.eu"} {
		res, body := get(host, "/")
		if res.StatusCode != 200 {
			t.Fatalf("%s: %d", host, res.StatusCode)
		}
		refs := ref.FindAllSubmatch(body, -1)
		if len(refs) == 0 {
			t.Errorf("%s: no image references", host)
		}
		for _, m := range refs {
			url := string(m[1])
			hm := hashed.FindStringSubmatch(url)
			if hm == nil {
				t.Errorf("%s: %s is not a hashed image URL", host, url)
				continue
			}
			want, err := images.ReadFile(hm[1] + "/" + hm[2] + hm[3])
			if err != nil {
				t.Errorf("%s: %s: %v", host, url, err)
				continue
			}
			res, b := get(host, url)
			if res.StatusCode != 200 || !bytes.Equal(b, want) {
				t.Errorf("%s: %d, or not the embedded bytes", url, res.StatusCode)
			}
			if cc := res.Header.Get("Cache-Control"); cc != "public, max-age=31536000, immutable" {
				t.Errorf("%s: Cache-Control %q", url, cc)
			}
			plain := "/" + hm[1] + "/" + hm[2] + hm[3]
			res, _ = get(host, plain)
			if res.StatusCode != 200 || res.Header.Get("Cache-Control") != "public, max-age=300" {
				t.Errorf("%s: %d, Cache-Control %q", plain, res.StatusCode, res.Header.Get("Cache-Control"))
			}
			wrong := "/" + hm[1] + "/" + hm[2] + ".00000000" + hm[3]
			if strings.Contains(url, ".00000000.") {
				wrong = "/" + hm[1] + "/" + hm[2] + ".11111111" + hm[3]
			}
			if res, _ = get(host, wrong); res.StatusCode != 404 {
				t.Errorf("%s: %d, want 404", wrong, res.StatusCode)
			}
		}
	}
}
