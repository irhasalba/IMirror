package main

import (
	"bufio"
	"io"
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
