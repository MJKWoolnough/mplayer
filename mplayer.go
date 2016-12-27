package mplayer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

var Executable string = "mplayer"

type MPlayer struct {
	cmd    *exec.Cmd
	stdin  io.Writer
	stdout *bufio.Reader
}

func Start(args ...string) (*MPlayer, error) {
	cmd := exec.Command(Executable, append([]string{"-slave", "-quiet", "-idle", "-input", "nodefault-bindings", "-noconfig", "all", "-msglevel", "all=-1:global=5"}, args...)...)
	cmd.Env = append(os.Environ(), "MPLAYER_VERBOSE=-1")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	return &MPlayer{
		cmd:    cmd,
		stdin:  stdin,
		stdout: bufio.NewReader(stdout),
	}, nil
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
