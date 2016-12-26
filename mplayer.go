package mplayer

import (
	"io"
	"os/exec"
)

var Executable string = "mplayer"

type MPlayer struct {
	cmd   *exec.Cmd
	stdin io.WriteCloser
}

func Start() (*MPlayer, error) {
	cmd := exec.Command(Executable, "-slave", "-quiet", "-idle", "-input", "nodefault-bindings", "-noconfig", "all")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	return &MPlayer{
		cmd:   cmd,
		stdin: stdin,
	}, nil
}

func (m *MPlayer) Stop() error {
	_, err := m.stdin.Write([]byte("quit\n"))
	if err != nil {
		return err
	}
	return m.cmd.Wait()
}
