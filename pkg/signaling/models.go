package signaling

import (
	"sync"

	"nhooyr.io/websocket"
)

type SignalingServer struct {
	lock        sync.Mutex
	communities map[string][]string
	macs        map[string]bool
	connections map[string]websocket.Conn
}

type SignalingClient struct{}
