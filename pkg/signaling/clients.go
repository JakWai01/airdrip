package signaling

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func NewSignalingClient() *SignalingClient {
	return &SignalingClient{}
}

// take flags for community and mac
func (s *SignalingClient) HandleConn(laddrKey string, communityKey string, macKey string) {
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

			// send offer
			byteArray, err := json.Marshal(Offer{Opcode: string(offer), Mac: partnerMac, Payload: "Hello World"})
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

			// send answer
			byteArray, err := json.Marshal(Answer{Opcode: string(answer), Mac: partnerMac, Payload: "Answer"})
			if err != nil {
				panic(err)
			}

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			break
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
