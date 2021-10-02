package signaling

var (
	exit       = make(chan struct{})
	exitServer = make(chan struct{})
)
