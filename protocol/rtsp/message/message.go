package message

const (
	Setup    = "SETUP"
	Record   = "RECORD"
	Play     = "PLAY"
	Pause    = "PAUSE"
	Teardown = "TEARDOWN"
	Describe = "DESCRIBE"
)

type Message string
