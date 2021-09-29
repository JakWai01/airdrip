package signaling

import (
	"sync"

	"nhooyr.io/websocket"
)

type SignalingServer struct {
	lock        sync.Mutex
	communities map[string][]websocket.Conn
	ready       map[string]int
}

type SignalingClient struct{}
