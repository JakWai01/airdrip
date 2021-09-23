package signaling

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"strings"
)

// This signaling protocol is heavily inspired by the weron project created by @pojntfx
// Take a look at the specification by clicking the following link: https://github.com/pojntfx/weron/blob/main/docs/signaling-protocol.txt#L12

func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		communities:    map[string][]string{},
		macs:           map[string]bool{},
		connections:    map[string]net.Conn{},
		candidateCache: []string{},
	}
}

// Method of the type SignalingServer
func (s *SignalingServer) HandleConn(c net.Conn) {

	fatal := make(chan error)

	go func() {
		for {
			// Read message from connection
			message, err := bufio.NewReader(c).ReadString('\n')
			if err != nil {
				fatal <- err

				return
			}

			fmt.Println(message)

			values := make(map[string]json.RawMessage)

			// Parse message
			err = json.Unmarshal([]byte(message), &values)
			if err != nil {
				panic(err)
			}

			switch Opcode(strings.ReplaceAll(string(values["opcode"]), "\"", "")) {
			case application:
				var opcode Application

				err := json.Unmarshal([]byte(message), &opcode)
				if err != nil {
					panic(err)
				}

				if _, ok := s.macs[opcode.Mac]; ok {
					// Send Rejection. That Mac is already contained
					byteArray, err := json.Marshal(Rejection{Opcode: string(rejection)})
					if err != nil {
						panic(err)
					}

					_, err = c.Write(byteArray)
					if err != nil {
						panic(err)
					}
					break
				}

				s.connections[opcode.Mac] = c

				// Check if community exists and if there are less than 2 members inside
				if val, ok := s.communities[opcode.Community]; ok {
					if len(val) >= 2 {
						// Send Rejection. This community is full
						byteArray, err := json.Marshal(Rejection{Opcode: string(rejection)})
						if err != nil {
							panic(err)
						}

						_, err = c.Write(byteArray)
						if err != nil {
							panic(err)
						}
						break
					} else {
						// Community exists but has less than 2 values in it
						s.communities[opcode.Community] = append(s.communities[opcode.Community], opcode.Mac)

						s.macs[opcode.Mac] = false

						byteArray, err := json.Marshal(Acceptance{Opcode: string(acceptance)})
						if err != nil {
							panic(err)
						}

						_, err = c.Write(byteArray)
						if err != nil {
							panic(err)
						}
						break
					}
				} else {
					// Community does not exist. Create community and insert mac
					s.communities[opcode.Community] = append(s.communities[opcode.Community], opcode.Mac)

					s.macs[opcode.Mac] = false

					byteArray, err := json.Marshal(Acceptance{Opcode: string(acceptance)})
					if err != nil {
						panic(err)
					}

					_, err = c.Write(byteArray)
					if err != nil {
						panic(err)
					}
					break
				}

			case ready:
				var opcode Ready

				err := json.Unmarshal([]byte(message), &opcode)
				if err != nil {
					panic(err)
				}

				// If we receive ready, mark the sending person as ready and check if both are ready. Loop through all communities to get the community the person is in.
				s.macs[opcode.Mac] = true

				// Loop through all members of the community and through all elements in it. If the mac isn't member of a community, this will panic.
				community, err := s.getCommunity(opcode.Mac)
				if err != nil {
					panic(err)
				}

				if len(s.communities[community]) == 2 {
					if s.macs[s.communities[community][0]] == true && s.macs[s.communities[community][1]] == true {
						// Send an introduction to the peer containing the address of the first peer.
						byteArray, err := json.Marshal(Introduction{Opcode: string(introduction), Mac: s.communities[community][0]})
						if err != nil {
							panic(err)
						}

						_, err = c.Write(byteArray)
						if err != nil {
							panic(err)
						}
						break

					}
				}

				break

			case offer:
				var opcode Offer

				err := json.Unmarshal([]byte(message), &opcode)
				if err != nil {
					panic(err)
				}

				// Get connection of the reveiver and send him the payload
				receiver := s.connections[opcode.Mac]

				var senderMac string
				// Get the Mac based on the current connection out of the connections mac
				for key, val := range s.connections {
					if c == val {
						senderMac = key
					}
				}

				byteArray, err := json.Marshal(Offer{Opcode: string(offer), Mac: senderMac, Payload: opcode.Payload})
				if err != nil {
					panic(err)
				}

				_, err = receiver.Write(byteArray)
				if err != nil {
					panic(err)
				}
				break
			case answer:
				var opcode Answer

				err := json.Unmarshal([]byte(message), &opcode)
				if err != nil {
					panic(err)
				}

				// Get connection of the receiver and send him the payload
				receiver := s.connections[opcode.Mac]

				var senderMac string
				// Get the Mac based on the current connection out of the connections mac
				for key, val := range s.connections {
					if c == val {
						senderMac = key
					}
				}

				byteArray, err := json.Marshal(Answer{Opcode: string(answer), Mac: senderMac, Payload: opcode.Payload})
				if err != nil {
					panic(err)
				}

				_, err = receiver.Write(byteArray)
				if err != nil {
					panic(err)
				}

				break
			case candidate:
				var opcode Candidate

				err := json.Unmarshal([]byte(message), &opcode)
				if err != nil {
					panic(err)
				}

				// Get connection of the receiver and send him the payload
				receiver := s.connections[opcode.Mac]

				var senderMac string
				// Get the Mac based on the current connection out of the connections mac
				for key, val := range s.connections {
					if c == val {
						senderMac = key
					}
				}

				byteArray, err := json.Marshal(Candidate{Opcode: string(candidate), Mac: senderMac, Payload: opcode.Payload})
				if err != nil {
					panic(err)
				}

				// Only write if we haven't written yet
				if contains(s.candidateCache, opcode.Mac) {
					break
				} else {
					s.candidateCache = append(s.candidateCache, opcode.Mac)

					_, err = receiver.Write(byteArray)
					if err != nil {
						panic(err)
					}
				}

				break

			case exited:
				var opcode Exited

				err := json.Unmarshal([]byte(message), &opcode)
				if err != nil {
					panic(err)
				}

				var senderMac string
				// Get the Mac based on the current connection out of the connections mac
				for key, val := range s.connections {
					if c == val {
						senderMac = key
					}
				}

				byteArray, err := json.Marshal(Resignation{Opcode: string(resignation), Mac: senderMac})
				if err != nil {
					panic(err)
				}
				var receiver net.Conn

				// Get the other peer in the community
				community, err := s.getCommunity(senderMac)
				if err != nil {
					panic(err)
				}

				if len(s.communities[community]) == 2 {
					if senderMac == s.communities[community][0] {
						// The second one is receiver
						receiver = s.connections[s.communities[community][1]]
					} else {
						// First one
						receiver = s.connections[s.communities[community][0]]
					}
				} else {
					receiver = s.connections[s.communities[community][0]]
				}

				// Send to the other peer
				_, err = receiver.Write(byteArray)
				if err != nil {
					panic(err)
				}

				// Remove this peer from all maps
				delete(s.macs, senderMac)
				delete(s.connections, senderMac)

				// Remove member from community
				s.communities[community] = deleteElement(s.communities[community], senderMac)

				// Remove community only if there is only one member left
				if len(s.communities[community]) == 0 {
					delete(s.communities, community)
				}

				return
			default:
				panic("Invalid message. Please use a valid opcode.")
			}
		}
	}()
}

func (s *SignalingServer) Close() []error {
	s.lock.Lock()
	defer s.lock.Unlock()

	errors := []error{}

	for _, conn := range s.connections {
		if err := conn.Close(); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
