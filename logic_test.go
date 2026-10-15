package main

import (
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestStatusFor(t *testing.T) {
	cases := map[string]state{
		"Connection request from 192.168.1.5": stateConnected,
		"client ACCEPTED":                     stateConnected,
		"Connection closed for socket 3":      stateWaiting,
		"RTSP TEARDOWN":                       stateWaiting,
		"Initialized server socket(s)":        stateNone,
	}
	for line, want := range cases {
		if got := statusFor(line); got != want {
			t.Errorf("statusFor(%q) = %v, want %v", line, got, want)
		}
	}
}

func TestUxplayCmd(t *testing.T) {
	exe, args, env := uxplayCmd(`C:\app\uxplay`, "Ruang Tamu", `C:\Windows`, `C:\Local`)
	if !strings.HasSuffix(exe, "uxplay.exe") {
		t.Errorf("exe = %q", exe)
	}
	if args[0] != "-n" || args[1] != "Ruang Tamu" || !slices.Contains(args, "d3d11videosink") {
		t.Errorf("args = %q", args)
	}
	if !strings.HasPrefix(env[0], "PATH=") || !strings.HasSuffix(env[0], `;C:\Windows`) {
		t.Errorf("PATH env = %q", env[0])
	}
}

func TestClipHandler(t *testing.T) {
	clip := "dari pc"
	h := clipHandler("rahasia", func() (string, error) { return clip, nil }, func(s string) error { clip = s; return nil })
	do := func(method, token, body string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(method, "/clip", strings.NewReader(body))
		r.Header.Set("X-Token", token)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		return w
	}
	if w := do("GET", "salah", ""); w.Code != 401 {
		t.Errorf("wrong token: code %d", w.Code)
	}
	if w := do("GET", "", ""); w.Code != 401 {
		t.Errorf("empty token: code %d", w.Code)
	}
	if w := do("POST", "salah", "jahat"); w.Code != 401 || clip != "dari pc" {
		t.Errorf("wrong token POST changed clip: code %d clip %q", w.Code, clip)
	}
	if w := do("GET", "rahasia", ""); w.Code != 200 || w.Body.String() != "dari pc" {
		t.Errorf("GET: code %d body %q", w.Code, w.Body)
	}
	if w := do("POST", "rahasia", "dari iphone"); w.Code != 204 || clip != "dari iphone" {
		t.Errorf("POST: code %d clip %q", w.Code, clip)
	}
	if w := do("POST", "rahasia", strings.Repeat("x", maxClip+1)); w.Code != 413 || clip != "dari iphone" {
		t.Errorf("oversize POST: code %d", w.Code)
	}
	if w := do("PUT", "rahasia", ""); w.Code != 405 {
		t.Errorf("PUT: code %d", w.Code)
	}
}

func TestLanguagesHaveSameKeys(t *testing.T) {
	for _, l := range langs {
		for k := range strs["en"] {
			if _, ok := strs[l.code][k]; !ok {
				t.Errorf("%s missing %q", l.code, k)
			}
		}
		if len(strs[l.code]) != len(strs["en"]) {
			t.Errorf("%s has %d keys, en has %d", l.code, len(strs[l.code]), len(strs["en"]))
		}
	}
}
