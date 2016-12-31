package mplayer

import (
	"bufio"
	"bytes"
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

var params = []string{"-slave", "-quiet", "-idle", "-input", "nodefault-bindings", "-noconfig", "all", "-msglevel", "all=-1:global=4:cfgparser=7"}

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
		pos:   -1,
	}

	go m.loop(bufio.NewReader(stdout))

	return m, nil
}

var (
	playListAdded = []byte("Adding file ")
	playListStart = []byte("Config pushed level is now 2")
	playListNext  = []byte("Config poped level=2")
	playListEnd   = []byte("Config poped level=1")
	response      = []byte("ANS_")
)

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
		if bytes.Equal(d, playListStart) {
			m.lock.Lock()
			m.pos = 0
			m.lock.Unlock()
		} else if bytes.Equal(d, playListNext) {
			m.lock.Lock()
			if m.pos == len(m.playlist) {
				m.pos = 0
			} else {
				m.pos++
			}
			m.lock.Unlock()
		} else if bytes.Equal(d, playListEnd) {
			m.lock.Lock()
			if m.loopAll {
				m.pos = 0
				m.startPlaylist()
			} else {
				m.pos = -1
			}
			if err := m.lock.Unlock(); err != nil {
				m.stdin.Write(quit)
				m.lock.Unlock()
				return
			}
		} else if bytes.HasPrefix(d, response) {

		} else if bytes.HasPrefix(d, playListAdded) {
			m.lock.Lock()
			m.playlist = append(m.playlist, string(d[len(playListAdded):]))
			m.lock.Unlock()
		}
	}
}

func (m *MPlayer) command(cmd []byte) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.err != nil {
		return m.err
	}
	_, m.err = m.stdin.Write(cmd)
	return m.err
}

func (m *MPlayer) query(query []byte, name string) (string, error) {
	return "", nil
}

func (m *MPlayer) startPlaylist() error {
	if m.err != nil {
		return m.err
	}
	var buf bytes.Buffer
	for n, file := range m.playlist {
		fmt.Fprintf(buf, "loadfile %q %d\n", file, n)
	}
	_, m.err = m.stdin.Write(buf.Bytes())
	return m.err
}

func (m *MPlayer) Quit() error {
	_, err := m.stdin.Write(quit)
	if err != nil {
		return err
	}
	return m.cmd.Wait()
}

func (m *MPlayer) Play(files ...string) error {
	m.lock.Lock()
	m.playlist = append(m.playlist[:0], files...)
	err := m.startPlaylist()
	m.lock.Unlock()
	return err
}

func (m *MPlayer) Next() error {
	return m.command(next)
}

func (m *MPlayer) Pause() error {
	return m.command(pause)
	return err
}

func (m *MPlayer) IsPaused() (bool, error) {
	ans, err := m.query(isPaused)
	if err != nil {
		return false, err
	}
	if ans == "yes\n" {
		return true, nil
	}
	return false, nil
}

func (m *MPlayer) Stop() error {
	return m.command(stop)
}
