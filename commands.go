package mplayer

import (
	"strconv"
	"strings"
	"time"

	"vimagination.zapto.org/errors"
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
	getVolume     = []byte("get_property volume\n")
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
	queryVolume

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

func (m *MPlayer) GetTrackLength() (time.Duration, error) {
	s, err := m.query(getLength, queryLength)
	if err != nil {
		return 0, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, err
	}
	return time.Duration(f * 1000 * 1000 * 1000), nil
}

func (m *MPlayer) Mute(on bool) error {
	if on {
		return m.command(muteOn)
	}
	return m.command(muteOff)
}

func (m *MPlayer) Muted() (bool, error) {
	s, err := m.query(isMuted, queryMuted)
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

func (m *MPlayer) GetVolume() (float64, error) {
	v, err := m.query(getVolume, queryVolume)
	if err != nil {
		return 0, err
	}
	return strconv.ParseFloat(v, 64)
}

func (m *MPlayer) Volume(v float64) error {
	return m.commandWrite("volume %.2f 1\n", v)
}

func (m *MPlayer) VolumeAdjust(v float64) error {
	return m.commandWrite("volume %.2f 0\n", v)
}

func (m *MPlayer) Seek(t time.Duration) error {
	return m.commandWrite("seek %.2f 0\n", float32(t)/1000/1000/1000)
}

func (m *MPlayer) SeekPercent(p float32) error {
	return m.commandWrite("seek %.2f 1\n", p)
}

func (m *MPlayer) SeekToTime(t time.Duration) error {
	return m.commandWrite("seek %.2f 2\n", float32(t)/1000/1000/1000)
}

func (m *MPlayer) GetPosition() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.pos
}

func (m *MPlayer) GetPlaylist() []string {
	m.lock.Lock()
	defer m.lock.Unlock()
	pl := make([]string, len(m.playlist))
	copy(pl, m.playlist)
	return pl
}

var (
	ErrInvalidResponse errors.Error = "mplayer returned an invalid response to the query"
)
