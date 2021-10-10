package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/JakWai01/airdrip/pkg/signaling"
)

func prettyJson(input string) (string, error) {
	var raw interface{}
	if err := json.Unmarshal([]byte(input), &raw); err != nil {
		return "", err
	}
	pretty, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return "", err
	}

	return string(pretty), nil
}

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

// func jsonWrapper() js.Func {
// 	jsonFunc := js.FuncOf(func(this js.Value, args []js.Value) interface{} {
// 		if len(args) != 1 {
// 			// When returning a value from Go to Javascript,
// 			// the ValueOf function will be used automatically by the compiler
// 			// to convert the Go value to a Javascript value.
// 			return "Invalid no. of arguments passed"
// 		}
// 		// This represents the first parameter passed from JavaScript.
// 		inputJSON := args[0].String()
// 		fmt.Printf("input %s\n", inputJSON)
// 		pretty, err := prettyJson(inputJSON)
// 		if err != nil {
// 			fmt.Printf("unable to convert to json %s\n", err)
// 			return err.Error()
// 		}
// 		return pretty
// 	})
// 	return jsonFunc
// }

func main() {
	fmt.Println("Go Web Assembly")
	// The jsonFunc which formats the JSON can be called from JavScript
	// using the function name formatJSON
	js.Global().Set("send", jsonWrapper())
	<-make(chan bool)
}
