package message

const (
	Setup    = "SETUP"
	Record   = "RECORD"
	Play     = "PLAY"
	Pause    = "PAUSE"
	Teardown = "TEARDOWN"
	Describe = "DESCRIBE"
	Exit     = "EXIT"
)

type Message string
