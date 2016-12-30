package mplayer

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"sync"
)

var Executable string = "mplayer"

type MPlayer struct {
	cmd   *exec.Cmd
	stdin io.Writer

	lock     sync.RWMutex
	err      error
	playlist []string
	pos      int
	loopAll  int
	replies  [numQueries][]chan []byte
}

func (r *responder) Write(p []byte)

var params = []string{"-slave", "-quiet", "-idle", "-input", "nodefault-bindings", "-noconfig", "all", "-msglevel", "all=-1:global=5:cfgparser=7 "}

func Start(args ...string) (*MPlayer, error) {
	cmd := exec.Command(Executable, append(params, args)...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	m := &MPlayer{
		cmd:   cmd,
		stdin: stdin,
	}

	br := bufio.NewReader(stdout)

	// read config parser lines to get initial playlist if any

	go m.loop(br)

	return m, nil
}

func (m *MPlayer) loop(stdout *bufio.Reader) {
	for {
		d, err := stdout.ReadBytes('\n')
		if err != nil {
			m.lock.Lock()
			m.err = err
			m.lock.Unlock()
			m.stdin.Write(quit)
			return
		}
		if len(d) < 4 { // not a line we care about
			continue
		}
		switch string(d[:4]) {
		case "EOF ":
			m.lock.Lock()
			if m.pos == len(m.playlist) {
				if m.loopAll != 0 {
					m.pos = 0
					// play list again
					if m.loopAll > 0 {
						m.loopAll--
					}
				} else {
					m.pos = -1
				}
			} else {
				m.pos++
			}
			m.lock.Unlock()
		case "ANS_":
			// response to query
		}
	}
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
