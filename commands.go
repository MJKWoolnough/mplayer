package mplayer

var (
	pause    = []byte("pause\n")
	stop     = []byte("stop\n")
	quit     = []byte("quit\n")
	isPaused = []byte("pausing_keep_force get_property pause\n")
	next     = []byte("pt_step 1 1\n")
)
