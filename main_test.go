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
		"/mcbot/acam-1-long-walk.webp":           200,
		"/mcbot/acam-1-long-walk-small.webp":     200,
		"/mcbot/acam-3-dusk-shelter.webp":        200,
		"/mcbot/acam-3-dusk-shelter-small.webp":  200,
		"/mcbot/acam-5-day2-morning.webp":        200,
		"/mcbot/acam-5-day2-morning-small.webp":  200,
		"/mcbot/acam-1-dusk-plains.webp":         404,
		"/mcbot/acam-1-dusk-plains-small.webp":   404,
		"/mcbot/acam-2-harvest-field.webp":       404,
		"/mcbot/acam-2-harvest-field-small.webp": 404,
		"/mcbot/acam-3-harvest-wheat.webp":       404,
		"/mcbot/acam-3-harvest-wheat-small.webp": 404,
		"/mcbot/dashboard.png":                   404, // the dashboard is private; no screenshot
		"/mcbot/dashboard-small.webp":            404,
		"/img/robot3d-pet.webp":                  200,
		"/img/robot3d-focus.webp":                200,
		"/img/robot3d-raw.webp":                  200,
		"/img/robot3d-dev.webp":                  200,
		"/img/robot3d-qr.webp":                   200,
		"/img/i41-robot.webp":                    404,
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
	for _, n := range []string{"acam-1-long-walk", "acam-3-dusk-shelter", "acam-5-day2-morning"} {
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

// The s.i41.eu card shows the same five robots as s.i41.eu, each linking to its app there.
func TestStackchanRobotRow(t *testing.T) {
	s := string(indexPage)
	if strings.Contains(s, "i41-robot.webp") {
		t.Error("index.html: still the single i41-robot picture")
	}
	for name, anchor := range map[string]string{"pet": "pet", "focus": "focus", "raw": "raw", "dev": "dev", "qr": "connect"} {
		re := regexp.MustCompile(`<a href="https://s\.i41\.eu/#` + anchor + `"[^>]*>\s*<figure><img src="/img/robot3d-` + name + `(?:\.[0-9a-f]{8})?\.webp"[^>]* alt="[^"]+"`)
		if !re.MatchString(s) {
			t.Errorf("index.html: no robot3d-%s.webp linking to s.i41.eu/#%s", name, anchor)
		}
	}
}

// On narrow screens only robots 1, 3 and 5 show, in one row of three.
func TestNarrowScreenThreeRobots(t *testing.T) {
	s := string(indexPage)
	i := strings.Index(s, "@media (max-width: 640px)")
	if i < 0 {
		t.Fatal("index.html: no max-width 640px media query")
	}
	m := s[i:]
	if j := strings.Index(m, "\n"); j >= 0 {
		m = m[:j]
	}
	for _, n := range []string{"2", "4"} {
		if !strings.Contains(m, ".robots a:nth-child("+n+")") {
			t.Errorf("index.html: narrow rule does not hide .robots a:nth-child(%s)", n)
		}
	}
	if !strings.Contains(m, "repeat(3, 1fr)") {
		t.Error("index.html: narrow robots row is not 3 columns")
	}
}

func TestMcbotViewerShots(t *testing.T) {
	s := string(mcbotPage)
	if !strings.Contains(s, "drawn by our own viewer, not the official Minecraft client, with textures of our own") {
		t.Error("mcbot.html: missing the sentence on our own viewer")
	}
	for _, name := range []string{"acam-1-long-walk", "acam-3-dusk-shelter", "acam-5-day2-morning"} {
		re := regexp.MustCompile(`<img src="/mcbot/` + name + `[^"]*\.webp"[^>]*>`)
		found := false
		for _, tag := range re.FindAllString(s, -1) {
			if strings.Contains(tag, name+"-small") {
				continue
			}
			found = true
			if !strings.Contains(tag, `width="1600" height="900"`) {
				t.Errorf("%s: full-size img is not 1600x900: %s", name, tag)
			}
		}
		if !found {
			t.Errorf("%s: no full-size img", name)
		}
	}
}
