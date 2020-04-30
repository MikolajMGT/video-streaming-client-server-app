package state

const (
	// only for internal thread management - not protocol compatible
	Detached = -1

	// protocol states
	Init = iota
	Ready
	Playing
	Recording
)

type State int
