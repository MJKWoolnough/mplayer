package mplayer

import (
	"strconv"
	"strings"

	"github.com/MJKWoolnough/errors"
)

var (
	pause         = []byte("pause\n")
	stop          = []byte("stop\n")
	quit          = []byte("quit\n")
	isPaused      = []byte("pausing_keep_force get_property pause\n")
	next          = []byte("pt_step 1 1\n")
	prev          = []byte("pt_step -1 1\n")
	getAlbum      = []byte("get_meta_album\n")
	getArtist     = []byte("get_meta_artist\n")
	getComment    = []byte("get_meta_comment\n")
	getGenre      = []byte("get_meta_genre\n")
	getTitle      = []byte("get_meta_title\n")
	getTrack      = []byte("get_meta_track\n")
	getYear       = []byte("get_meta_year\n")
	getLength     = []byte("get_time_length\n")
	fullscreenOn  = []byte("vo_fullscreen 1\n")
	fullscreenOff = []byte("vo_fullscreen 0\n")
	muteOn        = []byte("mute 1\n")
	muteOff       = []byte("mute 0\n")
	isMuted       = []byte("get_property mute\n")
)

const (
	queryPause = iota
	queryAlbum
	queryArtist
	queryComment
	queryGenre
	queryTitle
	queryTrack
	queryYear
	queryLength
	queryMuted

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

func (m *MPlayer) Prev() error {
	return m.command(prev)
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

func (m *MPlayer) GetAlbum() (string, error) {
	return m.getMD(getAlbum, queryAlbum)
}

func (m *MPlayer) GetArtist() (string, error) {
	return m.getMD(getArtist, queryArtist)
}

func (m *MPlayer) GetComment() (string, error) {
	return m.getMD(getComment, queryComment)
}

func (m *MPlayer) GetGenre() (string, error) {
	return m.getMD(getGenre, queryGenre)
}

func (m *MPlayer) GetTitle() (string, error) {
	return m.getMD(getTitle, queryTitle)
}

func (m *MPlayer) GetTrack() (string, error) {
	return m.getMD(getTrack, queryTrack)
}

func (m *MPlayer) GetYear() (string, error) {
	return m.getMD(getYear, queryYear)
}

func (m *MPlayer) getMD(q []byte, r int) (string, error) {
	s, err := m.query(getAlbum, queryAlbum)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(s[1 : len(s)-1]), nil
}

func (m *MPlayer) Fullscreen(full bool) error {
	if full {
		return m.command(fullscreenOn)
	}
	return m.command(fullscreenOff)
}

func (m *MPlayer) GetTrackLength() (float64, error) {
	s, err := m.query(getLength, queryLength)
	if err != nil {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}

func (m *MPlayer) Mute(on bool) error {
	if on {
		return m.command(muteOn)
	}
	return m.command(muteOff)
}

func (m *MPlayer) Muted() (bool, error) {
	s, err := s.query(isMuted, queryMuted)
	if err != nil {
		return false, err
	}
	if s == "on" {
		return true, nil
	} else if s == "off" {
		return false, nil
	}
	return false, ErrInvalidResponse
}

var (
	ErrInvalidResponse errors.Error = "mplayer returned an invalid response to the query"
)
