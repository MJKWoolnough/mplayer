package mplayer

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

var Executable string = "mplayer"

type MPlayer struct {
	in    chan<- []byte
	out   <-chan string
	close chan struct{}
}

var params = []string{"-slave", "-quiet", "-idle", "-input", "nodefault-bindings", "-noconfig", "all", "-msglevel", "all=-1:global=5"}

func Start(args ...string) *MPlayer {
	in := make(chan []byte)
	out := make(chan string)
	c := make(chan struct{})
	go start(append(params, args...), in, out, c)
	m := &MPlayer{
		in:    in,
		out:   out,
		close: c,
	}
	runtime.SetFinalizer(m, (*MPlayer).quit)
	return m
}

func start(args []string, in <-chan []byte, out chan<- string, c <-chan struct{}) {
	mplayer := exec.Command(Executable, args...)
	env := append(os.Environ(), "MPLAYER_VERBOSE=-1")
	mplayer.Env = env
	stdin, _ := mplayer.StdinPipe()
	stdout, _ := mplayer.StdoutPipe()
	mplayer.Start()

	for {
		select {
		case cmd := <-in:
		case <-c:
			close(out)
			stdin.Write(quit)
			mplayer.Wait()
			return
		}
	}
}

func (m *MPlayer) quit() {
	close(m.in)
	close(m.close)
}

func (m *MPlayer) play() error {
	return nil
}

func (m *MPlayer) command(cmd []byte) string {
	m.in <- cmd
	return <-m.out
}

func (m *MPlayer) Quit() error {
	_, err := m.stdin.Write(quit)
	if err != nil {
		return err
	}
	return m.cmd.Wait()
}

func (m *MPlayer) Play(files ...string) error {
	for n, file := range files {
		if _, err := fmt.Fprintf(m.stdin, "loadfile %q %d\n", file, n); err != nil {
			return err
		}
	}
	return nil
}

func (m *MPlayer) Next() error {
	_, err := m.stdin.Write(next)
	return err
}

func (m *MPlayer) Pause() error {
	_, err := m.stdin.Write(pause)
	return err
}

func (m *MPlayer) IsPaused() (bool, error) {
	_, err := m.stdin.Write(isPaused)
	if err != nil {
		return false, err
	}
	ans, err := m.stdout.ReadBytes('\n')
	if err != nil {
		return false, err
	}
	if string(ans) == "ANS_pause=yes\n" {
		return true, nil
	}
	return false, nil
}

func (m *MPlayer) Stop() error {
	_, err := m.stdin.Write(stop)
	return err
}
