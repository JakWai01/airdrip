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

// The whole point is to establish a WebRTC connection.

// take flags for community and mac
func (s *SignalingClient) HandleConn(laddrKey string, communityKey string, macKey string) {
	// Prepare the configuration
	var config = webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
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

	// send offer to other client

	// We need to receive the offer from the signaling server (how to generate an offer)

	// Then set remote description

	// Create answer

	// ...

	var partnerMac string

	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		panic(err)
	}

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

		// send exited
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

	// enter loop here
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
			// offer := webrtc.SessionDescription{}
			offer_var, err := peerConnection.CreateOffer(nil)
			if err != nil {
				panic(err)
			}

			// fmt.Println(offer_var.SDP)
			if err := peerConnection.SetLocalDescription(offer_var); err != nil {
				panic(err)
			}

			// send offer
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

			// Create channel that is blocked until ICE Gathering is complete
			gatherComplete := webrtc.GatheringCompletePromise(peerConnection)

			err = peerConnection.SetLocalDescription(answer_val)
			if err != nil {
				panic(err)
			}

			// send answer
			byteArray, err := json.Marshal(Answer{Opcode: string(answer), Mac: partnerMac, Payload: answer_val.SDP})
			if err != nil {
				panic(err)
			}

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			// Block until ICE Gathering is complete, disabling trickle ICE
			// we do this because we only can exchange one signaling message
			// in a production application you should exchange ICE Candidates via OnICECandidate
			<-gatherComplete

			// Output the answer in base64, so we can paste it in the browser
			fmt.Println(*peerConnection.LocalDescription())

			select {}

			// break
		case answer:
			var opcode Answer

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			partnerMac = opcode.Mac

			// send candidate
			byteArray, err := json.Marshal(Candidate{Opcode: string(candidate), Mac: partnerMac, Payload: "Candidate"})
			if err != nil {
				panic(err)
			}

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
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

			partnerMac = opcode.Mac

			// send candidate back
			byteArray, err := json.Marshal(Candidate{Opcode: string(candidate), Mac: partnerMac, Payload: "Candidate 2"})
			if err != nil {
				panic(err)
			}

			byteArray = append(byteArray, "\n"...)

			_, err = conn.Write([]byte(byteArray))
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
