package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

const createNoWindow = 0x08000000

type app struct {
	mw     *walk.MainWindow
	name   *walk.LineEdit
	btn    *walk.PushButton
	status *walk.Label
	log    *walk.TextEdit
	proc   *exec.Cmd
	bundle string
}

func main() {
	exePath, _ := os.Executable()
	a := &app{bundle: filepath.Join(filepath.Dir(exePath), "uxplay")}
	err := MainWindow{
		AssignTo: &a.mw,
		Title:    "iMirror",
		Size:     Size{Width: 620, Height: 420},
		Layout:   VBox{},
		Children: []Widget{
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					Label{Text: "Nama di iPhone:"},
					LineEdit{AssignTo: &a.name, Text: "iMirror", MaxSize: Size{Width: 150}},
					PushButton{AssignTo: &a.btn, Text: "Start", OnClicked: a.toggle},
					Label{AssignTo: &a.status, Text: "Stopped", TextColor: walk.RGB(128, 128, 128)},
					HSpacer{},
				},
			},
			TextEdit{AssignTo: &a.log, ReadOnly: true, VScroll: true, Font: Font{Family: "Consolas", PointSize: 9}},
		},
	}.Create()
	if err != nil {
		panic(err)
	}
	a.write("iPhone: Control Center > Screen Mirroring > pilih nama di atas.\n")
	for _, c := range a.doctor() {
		a.write(map[bool]string{true: "[OK]   ", false: "[MISS] "}[c.ok] + c.msg + "\n")
	}
	a.startClip()
	a.mw.Closing().Attach(func(*bool, walk.CloseReason) { a.kill() })
	a.mw.Run()
}

type check struct {
	ok  bool
	msg string
}

func (a *app) doctor() []check {
	exe, _, _ := uxplayCmd(a.bundle, "x", "", "")
	_, statErr := os.Stat(exe)
	sc := exec.Command("sc", "query", "Bonjour Service")
	sc.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	out, _ := sc.Output()
	return []check{
		{statErr == nil, `uxplay.exe ada di folder uxplay\bin`},
		{strings.Contains(string(out), "RUNNING"), "Bonjour Service berjalan (install 'Bonjour Print Services for Windows' dari Apple)"},
		{true, "Saat pertama Start, izinkan iMirror/uxplay di popup Windows Firewall (Private network)"},
	}
}

func (a *app) write(s string) {
	a.log.AppendText(strings.ReplaceAll(s, "\n", "\r\n"))
}

func (a *app) setStatus(text string, c walk.Color) {
	a.status.SetText(text)
	a.status.SetTextColor(c)
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
	name := strings.TrimSpace(a.name.Text())
	if name == "" {
		name = "iMirror"
	}
	exe, args, env := uxplayCmd(a.bundle, name, os.Getenv("PATH"), os.Getenv("LOCALAPPDATA"))
	if _, err := os.Stat(exe); err != nil {
		a.write("uxplay tidak ditemukan, lihat [MISS] di atas.\n")
		return
	}
	a.write("$ " + exe + " " + strings.Join(args, " ") + "\n")
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(), env...)
	cmd.SysProcAttr = uxplayProcAttr()
	pr, pw := io.Pipe()
	cmd.Stdout, cmd.Stderr = pw, pw
	if err := cmd.Start(); err != nil {
		a.write("gagal start: " + err.Error() + "\n")
		return
	}
	a.proc = cmd
	a.btn.SetText("Stop")
	a.setStatus("Menunggu iPhone...", walk.RGB(230, 140, 0))
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
		a.write("uxplay berhenti (exit " + cmd.ProcessState.String() + ")\n")
		a.proc = nil
		a.btn.SetText("Start")
		a.setStatus("Stopped", walk.RGB(128, 128, 128))
	})
}

func (a *app) onLine(line string) {
	a.write(line + "\n")
	switch statusFor(line) {
	case stateConnected:
		a.setStatus("Terhubung", walk.RGB(0, 150, 0))
	case stateWaiting:
		a.setStatus("Menunggu iPhone...", walk.RGB(230, 140, 0))
	}
}

// startClip serves clipHandler on the LAN; see logic.go for the iOS Shortcut side.
func (a *app) startClip() {
	token, err := clipToken(filepath.Join(os.Getenv("LOCALAPPDATA"), "iMirror-clip-token.txt"))
	if err != nil {
		a.write("Clipboard sync mati: " + err.Error() + "\n")
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
			a.mw.Synchronize(func() { a.write("Clipboard diterima dari iPhone\n") })
		}
		return err
	}
	a.write("Clipboard sync: http://" + lanIP() + ":" + clipPort + "/clip   header X-Token: " + token + "\n")
	go func() {
		err := http.ListenAndServe(":"+clipPort, clipHandler(token, get, set))
		a.mw.Synchronize(func() { a.write("Clipboard sync berhenti: " + err.Error() + "\n") })
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
