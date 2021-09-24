package signaling

import (
	"net"
	"sync"
)

type Opcode string

const (
	application  Opcode = "application"
	acceptance   Opcode = "acceptance"
	rejection    Opcode = "rejection"
	ready        Opcode = "ready"
	introduction Opcode = "introduction"
	offer        Opcode = "offer"
	answer       Opcode = "answer"
	candidate    Opcode = "candidate"
	exited       Opcode = "exited"
	resignation  Opcode = "resignation"
)

// type Application struct {
// 	Opcode    string `json:"opcode"`
// 	Community string `json:"community"`
// 	Mac       string `json:"mac"`
// }

// type Acceptance struct {
// 	Opcode string `json:"opcode"`
// }

// type Rejection struct {
// 	Opcode string `json:"opcode"`
// }

// type Ready struct {
// 	Opcode string `json:"opcode"`
// 	Mac    string `json:"mac"`
// }

// type Introduction struct {
// 	Opcode string `json:"opcode"`
// 	Mac    string `json:"mac"`
// }

// type Offer struct {
// 	Opcode  string `json:"opcode"`
// 	Mac     string `json:"mac"`
// 	Payload string `json:"payload"`
// }

// type Answer struct {
// 	Opcode  string `json:"opcode"`
// 	Mac     string `json:"mac"`
// 	Payload string `json:"payload"`
// }

// type Candidate struct {
// 	Opcode  string `json:"opcode"`
// 	Mac     string `json:"mac"`
// 	Payload string `json:"payload"`
// }

// type Exited struct {
// 	Opcode string `json:"opcode"`
// }

// type Resignation struct {
// 	Opcode string `json:"opcode"`
// 	Mac    string `json:"mac"`
// }

type SignalingServer struct {
	lock           sync.Mutex
	communities    map[string][]string
	macs           map[string]bool
	connections    map[string]net.Conn
	candidateCache []string
}

type SignalingClient struct{}
