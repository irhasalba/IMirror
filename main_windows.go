package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/lxn/walk"
)

const createNoWindow = 0x08000000

type app struct {
	mw       *walk.MainWindow
	name     *walk.LineEdit
	btn      *walk.CustomWidget
	btnFont  *walk.Font
	tile     *walk.CustomWidget
	iconFont *walk.Font
	dot      *walk.Label
	caption  *walk.Label
	title    *walk.Label
	sub      *walk.Label
	step3    *walk.Label
	endpoint *walk.LineEdit
	token    *walk.LineEdit
	reveal   *walk.Label
	clear    *walk.Label
	ready    *walk.Label
	log      *walk.TextEdit
	lang     string          // "en" or "id", see i18n.go
	mode     string          // "stopped", "waiting" or "connected": i18n key prefix
	checks   []check         // doctor results, re-rendered on language change
	revealed bool            // token shown in clear text
	langMenu *walk.Action    // toolbar language dropdown
	langActs []**walk.Action // one per langs entry
	retext   []func()        // re-applies translated text after a language change
	proc     *exec.Cmd
	bundle   string
}

// Layout lives in ui_windows.go, colors and widget helpers in style_windows.go.
func main() {
	exePath, _ := os.Executable()
	a := &app{bundle: filepath.Join(filepath.Dir(exePath), "uxplay")}
	a.checks = a.doctor()
	if err := a.buildUI(); err != nil {
		panic(err)
	}
	a.startClip()
	a.mw.Closing().Attach(func(*bool, walk.CloseReason) { a.kill() })
	a.mw.Run()
}

// check holds i18n keys, not text, so it can be re-rendered in another language.
type check struct {
	ok         bool
	title, sub string
	mono       bool
}

func quiet(name string, args ...string) string {
	c := exec.Command(name, args...)
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	out, _ := c.Output()
	return string(out)
}

func (a *app) doctor() []check {
	exe, _, _ := uxplayCmd(a.bundle, "x", "", "")
	_, statErr := os.Stat(exe)
	bonjour := strings.Contains(quiet("sc", "query", "Bonjour Service"), "RUNNING")
	// Rule is added by the installer (admin mode); portable copies rely on the Windows popup.
	fw := strings.Contains(quiet("netsh", "advfirewall", "firewall", "show", "rule", "name=iMirror (uxplay)"), "iMirror (uxplay)")
	cs := []check{
		{statErr == nil, "uxplay.ok", "uxplay.sub", true},
		{bonjour, "bonjour.ok", "bonjour.ok.sub", false},
		{fw, "firewall", "firewall.ok.sub", false},
	}
	if statErr != nil {
		cs[0].title = "uxplay.miss"
	}
	if !bonjour {
		cs[1].title, cs[1].sub = "bonjour.miss", "bonjour.miss.sub"
	}
	if !fw {
		cs[2].sub = "firewall.miss.sub"
	}
	return cs
}

func (a *app) write(s string) {
	a.log.AppendText(strings.ReplaceAll(s, "\n", "\r\n"))
}

func (a *app) kill() {
	if a.proc != nil && a.proc.Process != nil {
		a.proc.Process.Kill()
	}
}

func (a *app) toggle() {
	if a.proc != nil {
		a.kill()
		return
	}
	exe, args, env := uxplayCmd(a.bundle, a.deviceName(), os.Getenv("PATH"), os.Getenv("LOCALAPPDATA"))
	if _, err := os.Stat(exe); err != nil {
		a.write("uxplay.exe not found, see System checks.\n")
		return
	}
	a.write("$ " + exe + " " + strings.Join(args, " ") + "\n")
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.SysProcAttr = uxplayProcAttr()
	pr, pw := io.Pipe()
	cmd.Stdout, cmd.Stderr = pw, pw
	if err := cmd.Start(); err != nil {
		a.write("failed to start uxplay: " + err.Error() + "\n")
		return
	}
	a.proc = cmd
	a.show("waiting")
	go func() {
		err := cmd.Wait()
		pw.CloseWithError(err)
	}()
	go a.pump(cmd, pr)
}

// uxplayProcAttr hides uxplay's console only. Never HideWindow: it hides the video window too.
func uxplayProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: createNoWindow}
}

func (a *app) pump(cmd *exec.Cmd, r io.Reader) {
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		a.mw.Synchronize(func() { a.onLine(line) })
	}
	a.mw.Synchronize(func() {
		a.write("uxplay stopped (exit " + cmd.ProcessState.String() + ")\n")
		a.proc = nil
		a.show("stopped")
	})
}

func (a *app) onLine(line string) {
	a.write(line + "\n")
	switch statusFor(line) {
	case stateConnected:
		a.show("connected")
	case stateWaiting:
		a.show("waiting")
	}
}

// startClip serves clipHandler on the LAN; see logic.go for the iOS Shortcut side.
func (a *app) startClip() {
	token, err := clipToken(filepath.Join(os.Getenv("LOCALAPPDATA"), "iMirror-clip-token.txt"))
	if err != nil {
		a.endpoint.SetText(fmt.Sprintf(a.t("clip.off"), err))
		a.write("Clipboard sync off: " + err.Error() + "\n")
		return
	}
	// Win32 clipboard must be touched from the UI thread.
	onUI := func(f func() error) error {
		done := make(chan error, 1)
		a.mw.Synchronize(func() { done <- f() })
		return <-done
	}
	get := func() (text string, err error) {
		err = onUI(func() (e error) { text, e = walk.Clipboard().Text(); return })
		return
	}
	set := func(text string) error {
		err := onUI(func() error { return walk.Clipboard().SetText(text) })
		if err == nil {
			a.mw.Synchronize(func() { a.write("Clipboard received from iPhone\n") })
		}
		return err
	}
	a.endpoint.SetText("http://" + lanIP() + ":" + clipPort + "/clip")
	a.token.SetText(token)
	go func() {
		err := http.ListenAndServe(":"+clipPort, clipHandler(token, get, set))
		a.mw.Synchronize(func() { a.write("Clipboard sync stopped: " + err.Error() + "\n") })
	}()
}

// clipToken reads the saved token or creates one, so the iPhone Shortcut keeps working across restarts.
func clipToken(path string) (string, error) {
	if b, err := os.ReadFile(path); err == nil && len(strings.TrimSpace(string(b))) >= 16 {
		return strings.TrimSpace(string(b)), nil
	}
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := hex.EncodeToString(b)
	return token, os.WriteFile(path, []byte(token), 0o600)
}

// lanIP is the address the default route uses; UDP dial sends no packets.
func lanIP() string {
	c, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "<ip-pc>"
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).IP.String()
}
