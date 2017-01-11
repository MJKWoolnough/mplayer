package mplayer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"sync"

	"github.com/MJKWoolnough/errors"
)

var Executable string = "mplayer"

type MPlayer struct {
	cmd *exec.Cmd

	lock     sync.RWMutex
	stdin    io.Writer
	err      error
	playlist []string
	pos      int
	loopAll  int
	replies  [numQueries]chan<- string
}

var params = []string{"-slave", "-quiet", "-idle", "-input", "nodefault-bindings", "-noconfig", "all", "-msglevel", "all=-1:global=4:cfgparser=7"}

func Start(args ...string) (*MPlayer, error) {
	cmd := exec.Command(Executable, append(params, args...)...)
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
		cmd:     cmd,
		stdin:   stdin,
		pos:     -1,
		loopAll: -1,
	}

	_, err = stdin.Write(isPaused)
	if err != nil {
		cmd.Process.Kill()
		return nil, ErrInvalidStdin
	}

	br := bufio.NewReader(stdout)
	for {
		d, err := br.ReadBytes('\n')
		if err != nil {
			cmd.Process.Kill()
			return nil, ErrInvalidStdout
		}
		if bytes.HasPrefix(d, playListAdded) {
			m.playlist = append(m.playlist, string(d[len(playListAdded):]))
		} else if bytes.Equal(d, configParsed) {
			break
		}
	}
	go m.loop(br)

	return m, nil
}

var (
	configParsed  = []byte("ANS_pause=no\n")
	playListAdded = []byte("Adding file ")
	playListStart = []byte("Config pushed level is now 2\n")
	playListNext  = []byte("Config poped level=2\n")
	playListEnd   = []byte("Config poped level=1\n")
	response      = []byte("ANS_")
)

func (m *MPlayer) loop(stdout *bufio.Reader) {
	for {
		d, err := stdout.ReadBytes('\n')
		if err != nil {
			m.shutdown(err)
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
			if m.loopAll >= 0 {
				if m.loopAll > 0 {
					m.loopAll--
					if m.loopAll == 0 {
						m.loopAll = -1
					}
				}
				m.pos = 0
				m.startPlaylist()
			} else {
				m.pos = -1
			}
			m.lock.Unlock()
		} else if bytes.HasPrefix(d, response) {
			split := bytes.IndexByte(d, '=')
			if split == -1 {
				continue
			}
			responseType := -1
			switch string(d[len(response):split]) {
			case "pause":
				responseType = queryPause
			case "META_ALBUM":
				responseType = queryAlbum
			case "META_ARTIST":
				responseType = queryArtist
			case "META_COMMENT":
				responseType = queryComment
			case "META_GENRE":
				responseType = queryGenre
			case "META_TITLE":
				responseType = queryTitle
			case "META_TRACK":
				responseType = queryTrack
			case "META_YEAR":
				responseType = queryYear
			case "LENGTH":
				responseType = queryLength
			case "mute":
				responeType = queryMute
			}
			if responseType == -1 {
				continue
			}
			m.lock.Lock()
			rc := m.replies[responseType]
			m.replies[responseType] = nil
			m.lock.Unlock()
			rc <- string(d[split+1 : len(d)-1])
			close(rc)
		}
	}
}

func (m *MPlayer) shutdown(err error) error {
	m.lock.Lock()
	if err != nil {
		m.err = err
	}
	m.stdin.Write(quit)
	for _, ch := range m.replies {
		if ch != nil {
			close(ch)
		}
	}
	m.lock.Unlock()
	return m.cmd.Wait()
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

func (m *MPlayer) commandFunc(fn func(io.Writer) error) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	if m.err != nil {
		return m.err
	}
	m.err = fn(m.stdin)
	return m.err
}

func (m *MPlayer) query(query []byte, responseType int) (string, error) {
	m.lock.Lock()
	if m.err != nil {
		err := m.err
		m.lock.Unlock()
		return "", err
	}
	ch := m.replies[responseType]
	wc := make(chan string, 0)
	m.replies[responseType] = wc
	if ch == nil {
		_, m.err = m.stdin.Write(query)
	}
	m.lock.Unlock()
	data, ok := <-wc
	if !ok {
		if ch != nil {
			close(ch)
		}
		m.lock.RLock()
		err := m.err
		m.lock.RUnlock()
		return "", err
	}
	if ch != nil {
		ch <- data
		close(ch)
	}
	return data, nil
}

func (m *MPlayer) startPlaylist() error {
	if m.err != nil {
		return m.err
	}
	var buf bytes.Buffer
	for n, file := range m.playlist {
		fmt.Fprintf(&buf, "loadfile %q %d\n", file, n)
	}
	_, m.err = m.stdin.Write(buf.Bytes())
	return m.err
}

// Errors
const (
	ErrClosed        errors.Error = "closed"
	ErrInvalidStdin  errors.Error = "invalid stdin stream"
	ErrInvalidStdout errors.Error = "invalid stdout stream"
)
