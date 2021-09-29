package signaling

// Fix candidate exchange (Mac exchange)
// Neue Signaling Protocol ohne Mac ((nur conns pro community persistieren)
// Candidate Handling schritt fuer schritt

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	api "github.com/JakWai01/airdrip/pkg/api/websockets/v1"
	"github.com/pion/webrtc/v3"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
)

func NewSignalingClient() *SignalingClient {
	return &SignalingClient{}
}

func (s *SignalingClient) HandleConn(laddrKey string, communityKey string, macKey string) {
	conn, _, error := websocket.Dial(context.Background(), "ws://localhost:8080", nil)
	if error != nil {
		panic(error)
	}
	defer conn.Close(websocket.StatusInternalError, "the sky is falling")

	// Prepare configuration
	var config = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create RTCPeerConnection
	var peerConnection, err = webrtc.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}
	defer func() {
		if cErr := peerConnection.Close(); cErr != nil {
			log.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()

	var wg sync.WaitGroup
	wg.Add(1)

	candidates := make(chan string)

	var candidatesMux sync.Mutex

	// Introduce pending candidates. When a remote description is not set yet, the candidates will be cached
	// until a later invocation of the function
	pendingCandidates := make([]*webrtc.ICECandidate, 0)

	// Set the handler for peer connection state
	// This will notify you when the peer has connected/disconnected
	go func() {
		peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
			log.Printf("Peer Connection State has changed: %s\n", s.String())

			if s == webrtc.PeerConnectionStateFailed {
				// Wait until PeerConnection has had no network activity for 30 seconds or another failure.
				// Use webrtc.PeerCOnnectionStateDisconnected if you are interested in detecting faster timeout.
				// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
				log.Println("Peer Connection has gone to failed exiting")
				os.Exit(0)
			}
		})

		// This triggers when WE have a candidate for the other peer, not the other way around
		// This candidate key needs to be send to the other peer
		peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {

			if i == nil {
				return
			} else {
				// wg.Wait()
				candidatesMux.Lock()
				defer func() {
					candidatesMux.Unlock()
				}()

				desc := peerConnection.RemoteDescription()

				log.Println("SENDING SENDING")
				log.Println(b64.StdEncoding.EncodeToString([]byte(i.ToJSON().Candidate)))

				// data, err := json.Marshal(i)
				// if err != nil {
				// 	panic(err)
				// }

				if desc == nil {
					pendingCandidates = append(pendingCandidates, i)
					// Hier muss glaub ich die andere Mac gesendet werden
				} else if err := wsjson.Write(context.Background(), conn, api.NewCandidate(macKey, []byte(i.ToJSON().Candidate))); err != nil {
					panic(err)
				}
			}

		})

		// Register data channel creation handling
		peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {

			log.Println("OnDataChannel")
			// Register channel opening handling
			d.OnOpen(func() {
				log.Printf("Data channel '%s'-'%d' open. Messages will now be send to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())

				for range time.NewTicker(5 * time.Second).C {
					message := "Hello, World!"
					log.Printf("Sending '%s'\n", message)

					// Send the message as text
					sendErr := d.SendText(message)
					if sendErr != nil {
						panic(sendErr)
					}
				}
			})

			// Register text message handling
			d.OnMessage(func(msg webrtc.DataChannelMessage) {
				log.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
			})
		})

		if err := wsjson.Write(context.Background(), conn, api.NewApplication(communityKey, macKey)); err != nil {
			panic(err)
		}

		c := make(chan os.Signal)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c

			if err := wsjson.Write(context.Background(), conn, api.NewExited(macKey)); err != nil {
				panic(err)
			}

			os.Exit(0)
		}()

	}()

	go func() {
		for {
			// Read message from connection
			_, data, err := conn.Read(context.Background())
			if err != nil {
				panic(err)
			}

			log.Println(string(data))

			// Parse message
			var v api.Message
			if err := json.Unmarshal(data, &v); err != nil {
				panic(err)
			}

			// Handle different message types
			switch v.Opcode {
			case api.OpcodeAcceptance:
				if err := wsjson.Write(context.Background(), conn, api.NewReady(macKey)); err != nil {
					panic(err)
				}
				break
			case api.OpcodeIntroduction:
				// Create DataChannel
				sendChannel, err := peerConnection.CreateDataChannel("foo", nil)
				if err != nil {
					panic(err)
				}
				sendChannel.OnClose(func() {
					log.Println("sendChannel has closed")
				})
				sendChannel.OnOpen(func() {
					log.Println("sendChannel has opened")
				})
				sendChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
					log.Printf("Message from DataChannel %s payload %s", sendChannel.Label(), string(msg.Data))
				})

				var introduction api.Introduction
				if err := json.Unmarshal(data, &introduction); err != nil {
					panic(err)
				}

				partnerMac := introduction.Mac

				offer, err := peerConnection.CreateOffer(nil)
				if err != nil {
					panic(err)
				}

				if err := peerConnection.SetLocalDescription(offer); err != nil {
					panic(err)
				}

				fmt.Println("OFFER", offer)

				data, err := json.Marshal(offer)
				if err != nil {
					panic(err)
				}

				fmt.Println("SENDING", string(data))

				if err := wsjson.Write(context.Background(), conn, api.NewOffer(data, partnerMac)); err != nil {
					panic(err)
				}

				break
			case api.OpcodeOffer:
				var offer api.Offer
				if err := json.Unmarshal(data, &offer); err != nil {
					panic(err)
				}

				fmt.Println("GOT OFFER", string(offer.Payload))
				partnerMac := offer.Mac

				var offer_val webrtc.SessionDescription

				if err := json.Unmarshal([]byte(offer.Payload), &offer_val); err != nil {
					panic(err)
				}

				if err := peerConnection.SetRemoteDescription(offer_val); err != nil {
					panic(err)
				}

				go func() {
					for candidate := range candidates {
						log.Println("Candidate", candidate)

						if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate}); err != nil {
							panic(err)
						}
					}
				}()

				// Add pending candidates if there are any
				// if len(pendingCandidates) > 0 {
				// 	for _, candidate := range pendingCandidates {
				// 		if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate.ToJSON().Candidate}); err != nil {
				// 			panic(err)
				// 		}
				// 	}
				// }

				answer_val, err := peerConnection.CreateAnswer(nil)
				if err != nil {
					panic(err)
				}

				fmt.Println("SENDING", answer_val)
				err = peerConnection.SetLocalDescription(answer_val)
				if err != nil {
					panic(err)
				}

				data, err := json.Marshal(answer_val)
				if err != nil {
					panic(err)
				}

				// ALskdjlakjsdkla
				if err := wsjson.Write(context.Background(), conn, api.NewAnswer(data, partnerMac)); err != nil {
					panic(err)
				}

				wg.Done()
				break
			case api.OpcodeAnswer:
				var answer api.Answer
				if err := json.Unmarshal(data, &answer); err != nil {
					panic(err)
				}

				fmt.Println("GOT ANSWER", string(answer.Payload))
				// answer_val := webrtc.SessionDescription{}
				// answer_val.SDP = string(answer.Payload)
				// answer_val.Type = webrtc.SDPTypeAnswer
				var answer_val webrtc.SessionDescription

				if err := json.Unmarshal([]byte(answer.Payload), &answer_val); err != nil {
					panic(err)
				}

				if err := peerConnection.SetRemoteDescription(answer_val); err != nil {
					panic(err)
				}

				// Add pending candidates if there are any
				// if len(pendingCandidates) > 0 {
				// 	for _, candidate := range pendingCandidates {
				// 		if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: candidate.ToJSON().Candidate}); err != nil {
				// 			panic(err)
				// 		}
				// 	}
				// }

				wg.Done()
				break
			case api.OpcodeCandidate:
				var candidate api.Candidate
				if err := json.Unmarshal(data, &candidate); err != nil {
					panic(err)
				}

				// if peerConnection.RemoteDescription() != nil {
				// 	if err := peerConnection.AddICECandidate(webrtc.ICECandidateInit{Candidate: string(candidate.Payload)}); err != nil {
				// 		panic(err)
				// 	}
				// }

				go func() {
					candidates <- string(candidate.Payload)
				}()

				break
			case api.OpcodeResignation:
				if err := wsjson.Write(context.Background(), conn, api.NewExited(macKey)); err != nil {
					panic(err)
				}

				os.Exit(0)
			}
		}
	}()

	select {}

}
