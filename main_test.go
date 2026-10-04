package main

import (
	"bytes"
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
		"sm.w42.eu":        indexHTML,
	} {
		if !bytes.Equal(page(host), want) {
			t.Errorf("page(%q): wrong page", host)
		}
	}
}
