package signaling

import (
	"context"
	"encoding/json"
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
		log.Fatal(error)
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

	var candidatesMux sync.Mutex

	// Introduce pending candidates. When a remote description is not set yet, the candidates will be cached
	// until a later invocation of the function
	pendingCandidates := make([]*webrtc.ICECandidate, 0)

	// Set the handler for peer connection state
	// This will notify you when the peer has connected/disconnected
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
		}

		candidatesMux.Lock()
		defer candidatesMux.Unlock()

		desc := peerConnection.RemoteDescription()
		if desc == nil {
			pendingCandidates = append(pendingCandidates, i)
		} else if err := wsjson.Write(context.Background(), conn, api.NewCandidate(macKey, i.ToJSON().Candidate)); err != nil {
			log.Fatal(err)
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
					log.Fatal(sendErr)
				}
			}
		})

		// Register text message handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
	})

	// var partnerMac string

	if err := wsjson.Write(context.Background(), conn, api.NewApplication(communityKey, macKey)); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		if err := wsjson.Write(context.Background(), conn, api.NewExited(macKey)); err != nil {
			log.Fatal(err)
		}

		os.Exit(0)
	}()

	for {
		// Read message from connection
		_, data, err := conn.Read(context.Background())
		if err != nil {
			log.Fatal(err)
		}

		log.Println(string(data))

		// Parse message
		var v api.Message
		if err := json.Unmarshal(data, &v); err != nil {
			log.Fatal(err)
		}

		// Handle different message types
		switch v.Opcode {
		case api.OpcodeAcceptance:
			if err := wsjson.Write(context.Background(), conn, api.NewReady(macKey)); err != nil {
				log.Fatal(err)
			}
			break
		case api.OpcodeIntroduction:
			// Create DataChannel
			sendChannel, err := peerConnection.CreateDataChannel("foo", nil)
			if err != nil {
				log.Fatal(err)
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
				log.Fatal(err)
			}

			partnerMac := introduction.Mac

			offer, err := peerConnection.CreateOffer(nil)
			if err != nil {
				log.Fatal(err)
			}

			if err := peerConnection.SetLocalDescription(offer); err != nil {
				log.Fatal(err)
			}

			if err := wsjson.Write(context.Background(), conn, api.NewOffer(partnerMac, offer.SDP)); err != nil {
				log.Fatal(err)
			}

			break
		case api.OpcodeOffer:
			var offer api.Offer
			if err := json.Unmarshal(data, &offer); err != nil {
				log.Fatal(err)
			}

			partnerMac := offer.Mac

			offer_val := webrtc.SessionDescription{}
			offer_val.SDP = string(offer.Payload)
			offer_val.Type = webrtc.SDPTypeOffer

			if err := peerConnection.SetRemoteDescription(offer_val); err != nil {
				log.Fatal(err)
			}

			answer_val, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				log.Fatal(err)
			}

			err = peerConnection.SetLocalDescription(answer_val)
			if err != nil {
				log.Fatal(err)
			}

			if err := wsjson.Write(context.Background(), conn, api.NewAnswer(partnerMac, answer_val.SDP)); err != nil {
				log.Fatal(err)
			}

			wg.Done()
			break
		case api.OpcodeAnswer:
			var answer api.Answer
			if err := json.Unmarshal(data, &answer); err != nil {
				log.Fatal(err)
			}

			answer_val := webrtc.SessionDescription{}
			answer_val.SDP = string(answer.Payload)
			answer_val.Type = webrtc.SDPTypeAnswer

			if err := peerConnection.SetRemoteDescription(answer_val); err != nil {
				log.Fatal(err)
			}

			// Add pending candidates if there are any
			if len(pendingCandidates) > 0 {
				for _, candidate := range pendingCandidates {
					if err := peerConnection.AddICECandidate(candidate.ToJSON()); err != nil {
						log.Fatal(err)
					}
				}
			}

			wg.Done()
			break
		case api.OpcodeCandidate:
			var candidate api.Candidate
			if err := json.Unmarshal(data, &candidate); err != nil {
				log.Fatal(err)
			}

			candidate_val := webrtc.ICECandidateInit{}
			candidate_val.Candidate = string(candidate.Payload)

			if peerConnection.RemoteDescription() != nil {
				err = peerConnection.AddICECandidate(candidate_val)

				if err != nil {
					log.Fatal(err)
				}
			}

			break
		case api.OpcodeResignation:
			if err := wsjson.Write(context.Background(), conn, api.NewExited(macKey)); err != nil {
				log.Fatal(err)
			}

			os.Exit(0)
		}
	}
}
