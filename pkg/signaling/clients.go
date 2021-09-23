package signaling

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pion/webrtc/v3"
)

func NewSignalingClient() *SignalingClient {
	return &SignalingClient{}
}

func (s *SignalingClient) HandleConn(laddrKey string, communityKey string, macKey string) {

	conn, error := net.Dial("tcp", "localhost:8080")
	if error != nil {
		panic(error)
	}

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
			fmt.Printf("cannot close peerConnection: %v\n", cErr)
		}
	}()

	// Create DataChannel
	sendChannel, err := peerConnection.CreateDataChannel("foo", nil)
	if err != nil {
		panic(err)
	}
	sendChannel.OnClose(func() {
		fmt.Println("sendChannel has closed")
	})
	sendChannel.OnOpen(func() {
		fmt.Println("sendChannel has opened")
	})
	sendChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		fmt.Println(fmt.Sprintf("Message fromDatachannel %s payload %s", sendChannel.Label(), string(msg.Data)))
	})

	// Set the handler for Peer connection state
	// This will notify you when the peer has connected/disconnected
	peerConnection.OnConnectionStateChange(func(s webrtc.PeerConnectionState) {
		fmt.Printf("Peer Connection State has changed: %s\n", s.String())

		if s == webrtc.PeerConnectionStateFailed {
			// Wait until PeerConnection has had no network activity for 30 seconds or another failure.
			// Use webrtc.PeerConnectionStateDisconnected if you are interested in detecting faster timeout.
			// Note that the PeerConnection may come back from PeerConnectionStateDisconnected.
			fmt.Println("Peer Connection has gone to failed exiting")
			os.Exit(0)
		}
	})

	// This triggers when WE have a candidate for the other peer, not the other way around
	// This candidate key needs to be send to the other peer
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {

		// If nil isn't checked here, the program will throw a SEGFAULT at the end of conversation (as specified in: https://developer.mozilla.org/en-US/docs/Web/API/RTCPeerConnection/onicecandidate)
		if i != nil {
			candidate := Candidate{Opcode: "candidate", Mac: macKey, Payload: i.ToJSON().Candidate}
			fmt.Println(candidate.Payload)

			byteArray, err := json.Marshal(candidate)
			if err != nil {
				panic(err)
			}

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

		}
	})

	// ---------------------------------------------------------------------------------------
	// This is the information we get

	// Set ICE Candidate handler. As soon as a PeerConnection has gathered a candidate
	// send it to the other peer
	// answerPC.OnICECandidate(func(i *webrtc.ICECandidate) {
	// 	if i != nil {
	// 		check(offerPC.AddICECandidate(i.ToJSON()))
	// 	}
	// })
	// ---------------------------------------------------------------------------------------

	// Register data channel creation handling
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		fmt.Printf("New DataChannel %s %d\n", d.Label(), d.ID())

		// Register channel opening handling
		d.OnOpen(func() {
			fmt.Printf("Data channel '%s'-'%d' open. Messages will now be sent to any connected DataChannels every 5 seconds\n", d.Label(), d.ID())

			for range time.NewTicker(5 * time.Second).C {
				message := "Hello, World!"
				fmt.Printf("Sending '%s'\n", message)

				// Send the message as text
				sendErr := d.SendText(message)
				if sendErr != nil {
					panic(sendErr)
				}
			}
		})

		// Register text mesasage handling
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			fmt.Printf("Message from DataChannel '%s': '%s'\n", d.Label(), string(msg.Data))
		})
	})

	var partnerMac string

	application := Application{Opcode: "application", Community: communityKey, Mac: macKey}

	byteArray, err := json.Marshal(application)
	if err != nil {
		panic(err)
	}

	byteArray = append(byteArray, "\n"...)
	_, err = conn.Write([]byte(byteArray))
	if err != nil {
		panic(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c

		byteArray, err := json.Marshal(Exited{Opcode: string(exited)})
		if err != nil {
			panic(err)
		}

		byteArray = append(byteArray, "\n"...)

		_, err = conn.Write([]byte(byteArray))
		if err != nil {
			panic(err)
		}

		os.Exit(0)
	}()

	for {
		var input [1024]byte

		o, err := conn.Read(input[0:])
		if err != nil {
			os.Exit(0)
		}

		message := string(input[0:o])

		fmt.Println(message)

		values := make(map[string]json.RawMessage)

		err = json.Unmarshal([]byte(message), &values)
		if err != nil {
			panic(err)
		}

		switch Opcode(strings.ReplaceAll(string(values["opcode"]), "\"", "")) {
		case acceptance:

			// We actually don't need to unmarshal here because acceptance only contains an opcode
			byteArray, err := json.Marshal(Ready{Opcode: string(ready), Mac: macKey})
			if err != nil {
				panic(err)
			}

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			break
		case introduction:

			// We get the mac of our partner and store it
			var opcode Introduction

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			partnerMac = opcode.Mac

			// Wait for the offer to be pasted
			offer_var, err := peerConnection.CreateOffer(nil)
			if err != nil {
				panic(err)
			}

			if err := peerConnection.SetLocalDescription(offer_var); err != nil {
				panic(err)
			}

			byteArray, err := json.Marshal(Offer{Opcode: string(offer), Mac: partnerMac, Payload: offer_var.SDP})
			if err != nil {
				panic(err)
			}

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			break
		case offer:

			var opcode Offer

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			partnerMac = opcode.Mac

			payload := opcode.Payload

			fmt.Println(payload)

			offer_val := webrtc.SessionDescription{}
			offer_val.SDP = payload
			offer_val.Type = webrtc.SDPTypeOffer

			if err := peerConnection.SetRemoteDescription(offer_val); err != nil {
				panic(err)
			}

			answer_val, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				panic(err)
			}

			err = peerConnection.SetLocalDescription(answer_val)
			if err != nil {
				panic(err)
			}

			byteArray, err := json.Marshal(Answer{Opcode: string(answer), Mac: partnerMac, Payload: answer_val.SDP})
			if err != nil {
				panic(err)
			}

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			fmt.Println(*peerConnection.LocalDescription())

			select {}
		case answer:
			var opcode Answer

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			partnerMac = opcode.Mac

			offer_val := webrtc.SessionDescription{}
			offer_val.SDP = opcode.Payload
			offer_val.Type = webrtc.SDPTypeOffer

			if err := peerConnection.SetRemoteDescription(offer_val); err != nil {
				panic(err)
			}

			break
		case candidate:

			var opcode Candidate

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			candidate_val := webrtc.ICECandidateInit{}
			candidate_val.Candidate = opcode.Payload

			err = peerConnection.AddICECandidate(candidate_val)
			if err != nil {
				panic(err)
			}

			// check for candidates
			break
		case resignation:
			byteArray, err := json.Marshal(Exited{Opcode: string(exited)})
			if err != nil {
				panic(err)
			}

			byteArray = append(byteArray, "\n"...)

			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			os.Exit(0)
		}

	}
}
