package signaling

import (
	"context"
	"encoding/json"

	api "github.com/JakWai01/airdrip/pkg/api/websockets/v1"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// This signaling protocol is heavily inspired by the weron project created by @pojntfx
// Take a look at the specification by clicking the following link: https://github.com/pojntfx/weron/blob/main/docs/signaling-protocol.txt#L12

func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		communities: map[string][]string{},
		macs:        map[string]bool{},
		connections: map[string]websocket.Conn{},
	}
}

func (s *SignalingServer) HandleConn(conn websocket.Conn) {

	go func() {
		for {

			// Read message from connection
			_, data, err := conn.Read(context.Background())
			if err != nil {
				panic(err)
			}

			// Parse message
			var v api.Message
			if err := json.Unmarshal(data, &v); err != nil {
				panic(err)
			}

			// Handle different message types
			switch v.Opcode {
			case api.OpcodeApplication:
				var application api.Application
				if err := json.Unmarshal(data, &application); err != nil {
					panic(err)
				}

				if _, ok := s.macs[application.Mac]; ok {
					// Send rejection. That mac is already contained

					// Check if this conn is correct
					if err := wsjson.Write(context.Background(), &conn, api.NewRejection()); err != nil {
						panic(err)
					}
					break
				}

				s.connections[application.Mac] = conn

				// Check if community exists and if there are less than 2 members inside
				if val, ok := s.communities[application.Community]; ok {
					if len(val) >= 2 {
						// Send Rejection. This community is full
						if err := wsjson.Write(context.Background(), &conn, api.NewRejection()); err != nil {
							panic(err)
						}

						break
					} else {
						// Community exists and has less than 2 members inside
						s.communities[application.Community] = append(s.communities[application.Community], application.Mac)
						s.macs[application.Mac] = false

						if err := wsjson.Write(context.Background(), &conn, api.NewAcceptance()); err != nil {
							panic(err)
						}

						break
					}
				} else {
					// Community does not exist. Create community and insert mac
					s.communities[application.Community] = append(s.communities[application.Community], application.Mac)
					s.macs[application.Mac] = false

					if err := wsjson.Write(context.Background(), &conn, api.NewAcceptance()); err != nil {
						panic(err)
					}
					break
				}

			case api.OpcodeReady:
				var ready api.Ready
				if err := json.Unmarshal(data, &ready); err != nil {
					panic(err)
				}

				// If we receive ready, mark the sending person as ready and check if both are ready. Loop through all communities to get the community the person is in.
				s.macs[ready.Mac] = true

				// Loop thorugh all members of the community and thorugh all elements in it. If the mac isn't member of a community, this will panic.
				community, err := s.getCommunity(ready.Mac)
				if err != nil {
					panic(err)
				}

				if len(s.communities[community]) == 2 {
					if s.macs[s.communities[community][0]] == true && s.macs[s.communities[community][1]] == true {
						// Send an introduction to the peer containing the address of the first peer.
						if err := wsjson.Write(context.Background(), &conn, api.NewIntroduction(s.communities[community][0])); err != nil {
							panic(err)
						}
						break
					}
				}
				break
			case api.OpcodeOffer:
				var offer api.Offer
				if err := json.Unmarshal(data, &offer); err != nil {
					panic(err)
				}

				// Get the connection of the receiver and send him the payload
				receiver := s.connections[offer.Mac]

				community, err := s.getCommunity(offer.Mac)
				if err != nil {
					panic(err)
				}

				// We need to assign this
				offer.Mac = s.getSenderMac(offer.Mac, community)

				if err := wsjson.Write(context.Background(), &receiver, offer); err != nil {
					panic(err)
				}
				break
			case api.OpcodeAnswer:
				var answer api.Answer
				if err := json.Unmarshal(data, &answer); err != nil {
					panic(err)
				}

				// Get connection of the receiver and send him the payload
				receiver := s.connections[answer.Mac]

				community, err := s.getCommunity(answer.Mac)
				if err != nil {
					panic(err)
				}

				answer.Mac = s.getSenderMac(answer.Mac, community)

				if err := wsjson.Write(context.Background(), &receiver, answer); err != nil {
					panic(err)
				}

				break
			case api.OpcodeCandidate:
				var candidate api.Candidate
				if err := json.Unmarshal(data, &candidate); err != nil {
					panic(err)
				}

				community, err := s.getCommunity(candidate.Mac)
				if err != nil {
					panic(err)
				}

				candidate.Mac = s.getSenderMac(candidate.Mac, community)

				target := s.connections[candidate.Mac]

				if err := wsjson.Write(context.Background(), &target, candidate); err != nil {
					panic(err)
				}

				break
			case api.OpcodeExited:
				var exited api.Exited
				if err := json.Unmarshal(data, &exited); err != nil {
					panic(err)
				}

				var receiver websocket.Conn

				// Get the other peer in the community
				community, err := s.getCommunity(exited.Mac)
				if err != nil {
					panic(err)
				}

				if len(s.communities[community]) == 2 {
					if exited.Mac == s.communities[community][0] {
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
				if err := wsjson.Write(context.Background(), &receiver, api.NewResignation(exited.Mac)); err != nil {
					panic(err)
				}

				// Remove this peer from all maps
				delete(s.macs, exited.Mac)
				delete(s.connections, exited.Mac)

				// Remove meber from community
				s.communities[community] = deleteElement(s.communities[community], exited.Mac)

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

	for _, peer := range s.connections {
		if err := peer.Close(websocket.StatusGoingAway, "shutting down"); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}
