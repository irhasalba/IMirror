package main

import (
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
