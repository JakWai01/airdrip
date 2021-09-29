package signaling

import (
	"context"
	"encoding/json"
	"fmt"

	api "github.com/JakWai01/airdrip/pkg/api/websockets/v1"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func NewSignalingServer() *SignalingServer {
	return &SignalingServer{
		communities: map[string][]websocket.Conn{},
		// Optimally, this would be a map with websocket.Conn as key and bool as value
		// but since websocket.Conn can't be a key this is the improvised solution
		ready: map[string]int{},
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

			fmt.Println(string(data))

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

				if val, ok := s.communities[application.Community]; ok {
					if len(val) >= 2 {
						// Send rejection. This community is full

						if err := wsjson.Write(context.Background(), &conn, api.NewRejection()); err != nil {
							panic(err)
						}

						break
					} else {
						// Community exists and has less than 2 members inside
						s.communities[application.Community] = append(s.communities[application.Community], conn)

						if err := wsjson.Write(context.Background(), &conn, api.NewAcceptance()); err != nil {
							panic(err)
						}

						break
					}
				} else {
					// Community does not exist. Create community and insert mac
					s.communities[application.Community] = append(s.communities[application.Community], conn)

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

				// Get community based on conn
				community, err := s.getCommunity(conn)
				if err != nil {
					panic(err)
				}

				// If we receive ready, mark the sending person as ready
				s.ready[community] += 1

				if len(s.communities[community]) == 2 {
					if s.ready[community] == 2 {
						// Send introduction
						if err := wsjson.Write(context.Background(), &conn, api.NewIntroduction()); err != nil {
							panic(err)
						}
					}

					break
				}

				break
			case api.OpcodeOffer:
				var offer api.Offer
				if err := json.Unmarshal(data, &offer); err != nil {
					panic(err)
				}
			}
		}
	}()
}
