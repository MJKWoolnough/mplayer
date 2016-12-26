package mplayer

import "testing"

func TestStartStop(t *testing.T) {
	m, err := Start()
	if err != nil {
		t.Errorf("unexpected error while starting: %s", err)
		return
	}
	if err = m.Stop(); err != nil {
		t.Errorf("unexpected error while stopping: %s", err)
	}
}
