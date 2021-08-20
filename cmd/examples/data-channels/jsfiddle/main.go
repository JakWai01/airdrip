// +build js,wasm

package main

import (
	"fmt"

	"syscall/js"

	"examples/signal"

	"github.com/pion/webrtc/v3"
)

func main() {
	// Configure and create a new PeerConnection.
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	pc, err := webrtc.NewPeerConnection(config)
	if err != nil {
		handleError(err)
	}

	// Create a new channel (we can call the channel however we like)
	sendChannel, err := pc.CreateDataChannel("foo", nil)
	if err != nil {
		handleError(err)
	}

	// Log message on close
	sendChannel.OnClose(func() {
		fmt.Println("sendChannel has closed")
	})

	// Log message on open
	sendChannel.OnOpen(func() {
		fmt.Println("sendChannel has opened")
	})

	// Log message on message
	sendChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		log(fmt.Sprintf("Message from DataChannel %s payload %s", sendChannel.Label(), string(msg.Data)))
	})

	// Create offer
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		handleError(err)
	}

	// Set local description
	if err := pc.SetLocalDescription(offer); err != nil {
		handleError(err)
	}

	// Called when ICE connection state changes
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log(fmt.Sprint(state))
	})

	// Called on ICE candidate
	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			encodedDescr := signal.Encode(pc.LocalDescription())
			el := getElementByID("localSessionDescription")
			el.Set("value", encodedDescr)
		}
	})

	// Set up global callbacks which will be triggered on button clicks.
	js.Global().Set("sendMessage", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		go func() {
			el := getElementByID("message")
			message := el.Get("value").String()
			if message == "" {
				js.Global().Call("alert", "Message must not be empty")
				return
			}
			if err := sendChannel.SendText(message); err != nil {
				handleError(err)
			}
		}()
		return js.Undefined()
	}))

	// Implement startSession js function
	js.Global().Set("startSession", js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
		go func() {
			el := getElementByID("remoteSessionDescription")
			sd := el.Get("value").String()
			if sd == "" {
				js.Global().Call("alert", "Session Description must not be empty")
				return
			}

			// Session description interface
			descr := webrtc.SessionDescription{}
			// Store sd in descr
			signal.Decode(sd, &descr)

			// set the description of the other peer
			if err := pc.SetRemoteDescription(descr); err != nil {
				handleError(err)
			}
		}()
		return js.Undefined()
	}))

	select {}
}

// log messages to DOM
func log(msg string) {
	el := getElementByID("logs")
	el.Set("innerHTML", el.Get("innerHTML").String()+msg+"<br>")
}

func handleError(err error) {
	log("Unexpected error. Check console.")
	panic(err)
}

// implement getElementById
func getElementByID(id string) js.Value {
	return js.Global().Get("document").Call("getElementById", id)
}
