package main

import (
	"fmt"
	"syscall/js"

	"github.com/JakWai01/airdrip/pkg/signaling"
)

func startClient(community string, filename string, file []byte) {
	client := signaling.NewSignalingClient()
	go func() {
		client.HandleConn("airdrip.herokuapp.com", community, filename, file)
	}()
}

func jsonWrapper() js.Func {
	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		if len(args) != 3 {
			// When returning a value from Go to Javascript,
			// the ValueOf function will be used automatically by the compiler
			// to convert the Go value to a Javascript value.
			return "Invalid no. of arguments passed"
		}

		// The first parameter
		community := args[0].String()
		// The second parameter
		filename := args[1].String()
		// The third parameter
		file := args[2].String()

		startClient(community, filename, []byte(file))
		return "Client started"
	})
	return jsonFunc
}

func main() {
	fmt.Println("Go Web Assembly")
	js.Global().Set("send", jsonWrapper())
	<-make(chan bool)
}
