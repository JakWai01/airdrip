package main

import (
	"github.com/JakWai01/airdrip/cmd"
)

func main() {

	// This needs to be implemented on client side

	// config := webrtc.Configuration{
	// 	ICEServers: []webrtc.ICEServer{
	// 		{
	// 			URLs: []string{"stun:stun.l.google.com:19302"},
	// 		},
	// 	},
	// }

	// pc, err := webrtc.NewPeerConnection(config)
	// if err != nil {
	// 	panic(err)
	// }

	// offer, err := pc.CreateOffer(nil)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println(offer)

	// if err := pc.SetLocalDescription(offer); err != nil {
	// 	panic(err)
	// }

	// fmt.Println(pc.LocalDescription().Type.String())

	cmd.Execute()
}
