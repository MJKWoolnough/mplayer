package mplayer

import "github.com/MJKWoolnough/errors"

var (
	pause    = []byte("pause\n")
	stop     = []byte("stop\n")
	quit     = []byte("quit\n")
	isPaused = []byte("pausing_keep_force get_property pause\n")
	next     = []byte("pt_step 1 1\n")
)

const (
	queryPause = iota

	numQueries
)

func (m *MPlayer) Quit() error {
	return m.shutdown(ErrClosed)
}

func (m *MPlayer) Play(files ...string) error {
	m.lock.Lock()
	m.loopAll = -1
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
}

func (m *MPlayer) IsPaused() (bool, error) {
	ans, err := m.query(isPaused, queryPause)
	if err != nil {
		return false, err
	}
	if ans == "yes" {
		return true, nil
	} else if ans == "no" {
		return false, nil
	}
	return false, ErrInvalidResponse
}

func (m *MPlayer) Stop() error {
	return m.command(stop)
}

var (
	ErrInvalidResponse errors.Error = "mplayer returned an invalid response to the query"
)
