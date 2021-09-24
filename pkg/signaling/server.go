package signaling

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	api "github.com/JakWai01/airdrip/pkg/api/websockets/v1"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

// This signaling protocol is heavily inspired by the weron project created by @pojntfx
// Take a look at the specification by clicking the following link: https://github.com/pojntfx/weron/blob/main/docs/signaling-protocol.txt#L12

func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		communities:    map[string][]string{},
		macs:           map[string]bool{},
		connections:    map[string]websocket.Conn{},
		candidateCache: []string{},
	}
}

func (s *SignalingServer) HandleConn(conn websocket.Conn) {

	go func() {
		for {

			// Read message from connection
			_, data, err := conn.Read(context.Background())
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(string(data))

			// Parse message
			var v api.Message
			if err := json.Unmarshal(data, &v); err != nil {
				log.Fatal(err)
			}

			// Handle different message types
			switch v.Opcode {
			case api.OpcodeApplication:
				var application api.Application
				if err := json.Unmarshal(data, &application); err != nil {
					log.Fatal(err)
				}

				if _, ok := s.macs[application.Mac]; ok {
					fmt.Println("application")
					// Send rejection. That mac is already contained

					// Check if this conn is correct
					if err := wsjson.Write(context.Background(), &conn, api.NewRejection()); err != nil {
						log.Fatal(err)
					}
					break
				}

				s.connections[application.Mac] = conn

				// Check if community exists and if there are less than 2 members inside
				if val, ok := s.communities[application.Community]; ok {
					if len(val) >= 2 {
						// Send Rejection. This community is full
						if err := wsjson.Write(context.Background(), &conn, api.NewRejection()); err != nil {
							log.Fatal(err)
						}

						break
					} else {
						// Community exists and has less than 2 members inside
						s.communities[application.Community] = append(s.communities[application.Community], application.Mac)

						s.macs[application.Mac] = false

						if err := wsjson.Write(context.Background(), &conn, api.NewAcceptance()); err != nil {
							log.Fatal(err)
						}

						break
					}
				} else {
					// Community does not exist. Create community and insert mac
					s.communities[application.Community] = append(s.communities[application.Community], application.Mac)

					s.macs[application.Mac] = false

					if err := wsjson.Write(context.Background(), &conn, api.NewAcceptance()); err != nil {
						log.Fatal(err)
					}
					break
				}

			case api.OpcodeReady:
				var ready api.Ready
				if err := json.Unmarshal(data, &ready); err != nil {
					log.Fatal(err)
				}

				// If we receive ready, mark the sending person as ready and check if both are ready. Loop through all communities to get the community the person is in.
				s.macs[ready.Mac] = true

				// Loop thorugh all members of the community and thorugh all elements in it. If the mac isn't member of a community, this will panic.
				community, err := s.getCommunity(ready.Mac)
				if err != nil {
					log.Fatal(err)
				}

				if len(s.communities[community]) == 2 {
					if s.macs[s.communities[community][0]] == true && s.macs[s.communities[community][1]] == true {
						// Send an introduction to the peer containing the address of the first peer.
						if err := wsjson.Write(context.Background(), &conn, api.NewIntroduction(s.communities[community][0])); err != nil {
							log.Fatal(err)
						}
						break
					}
				}
				break
			case api.OpcodeOffer:
				var offer api.Offer
				if err := json.Unmarshal(data, &offer); err != nil {
					log.Fatal(err)
				}

				// Get the connection of the receiver and send him the payload
				receiver := s.connections[offer.Mac]

				community, err := s.getCommunity(offer.Mac)
				if err != nil {
					log.Fatal(err)
				}

				// We need to assign this
				var senderMac string

				if len(s.communities[community]) == 2 {
					if offer.Mac == s.communities[community][1] {
						// The second one is sender
						senderMac = s.communities[community][0]
					} else {
						// First one
						senderMac = s.communities[community][1]
					}
				} else {
					senderMac = s.communities[community][1]
				}

				if err := wsjson.Write(context.Background(), &receiver, api.NewOffer(senderMac, offer.Payload)); err != nil {
					log.Fatal(err)
				}
				break
			case api.OpcodeAnswer:
				var answer api.Answer
				if err := json.Unmarshal(data, &answer); err != nil {
					log.Fatal(err)
				}

				// Get connection of the receiver and send him the payload
				receiver := s.connections[answer.Mac]

				community, err := s.getCommunity(answer.Mac)
				if err != nil {
					log.Fatal(err)
				}

				var senderMac string

				if len(s.communities[community]) == 2 {
					if answer.Mac == s.communities[community][1] {
						// The second one is sender
						senderMac = s.communities[community][0]
					} else {
						// First one
						senderMac = s.communities[community][1]
					}
				} else {
					senderMac = s.communities[community][1]
				}

				if err := wsjson.Write(context.Background(), &receiver, api.NewAnswer(senderMac, answer.Payload)); err != nil {
					log.Fatal(err)
				}

				break
			case api.OpcodeCandidate:
			case api.OpcodeExited:
			}
		}
	}()
}
