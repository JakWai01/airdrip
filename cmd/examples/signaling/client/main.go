package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strings"
)

type Opcode string

const (
	application  Opcode = "application"
	acceptance   Opcode = "acceptance"
	rejection    Opcode = "rejection"
	ready        Opcode = "ready"
	introduction Opcode = "introduction"
	offer        Opcode = "offer"
	answer       Opcode = "answer"
	candidate    Opcode = "candidate"
	exited       Opcode = "exited"
	resignation  Opcode = "resignation"
)

type Application struct {
	Opcode    string `json:"opcode"`
	Community string `json:"community"`
	Mac       string `json:"mac"`
}

type Acceptance struct {
	Opcode string `json:"opcode"`
}

type Ready struct {
	Opcode string `json:"opcode"`
	Mac    string `json:"mac"`
}

type Offer struct {
	Opcode  string `json:"opcode"`
	Mac     string `json:"mac"`
	Payload string `json:"payload"`
}

type Answer struct {
	Opcode  string `json:"opcode"`
	Mac     string `json:"mac"`
	Payload string `json:"payload"`
}

type Introduction struct {
	Opcode string `json:"opcode"`
	Mac    string `json:"mac"`
}

type Candidate struct {
	Opcode  string `json:"opcode"`
	Mac     string `json:"mac"`
	Payload string `json:"payload"`
}

// take flags for community and mac
func main() {
	var laddr = flag.String("laddr", "localhost:8080", "listen address")
	var mac = flag.String("mac", "123", "mac (identification string)")
	var community = flag.String("community", "a", "community to join")
	flag.Parse()

	var partnerMac string

	conn, err := net.Dial("tcp", *laddr)
	if err != nil {
		panic(err)
	}

	application := Application{Opcode: "application", Community: *community, Mac: *mac}

	byteArray, err := json.Marshal(application)
	if err != nil {
		panic(err)
	}

	byteArray = append(byteArray, "\n"...)
	_, err = conn.Write([]byte(byteArray))
	if err != nil {
		panic(err)
	}

	// enter loop here
	for {
		var input [1024]byte

		o, err := conn.Read(input[0:])
		if err != nil {
			panic(err)
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
			fmt.Println("acceptance")

			// We actually don't need to unmarshal here because acceptance only contains an opcode
			byteArray, err := json.Marshal(Ready{Opcode: string(ready), Mac: *mac})
			if err != nil {
				panic(err)
			}

			fmt.Println(string(byteArray))

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			break
		case introduction:
			fmt.Println("introduction")

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

			fmt.Println(string(byteArray))

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			break
		case offer:
			fmt.Println("offer")

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

			fmt.Println(string(byteArray))

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			break
		case answer:
			fmt.Println("answer")

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

			fmt.Println(string(byteArray))

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			break
		case candidate:
			fmt.Println("candidate")

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

			fmt.Println(string(byteArray))

			byteArray = append(byteArray, "\n"...)
			_, err = conn.Write([]byte(byteArray))
			if err != nil {
				panic(err)
			}

			// check for candidates
			break
		}

	}

	// Ready
	// ready := Ready{Opcode: "ready", Mac: *mac}

	// byteArray, err = json.Marshal(ready)
	// if err != nil {
	// 	panic(err)
	// }

	// byteArray = append(byteArray, "\n"...)
	// _, err = conn.Write([]byte(byteArray))
	// if err != nil {
	// 	panic(err)
	// }

	// o, err = conn.Read(input[0:])
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println()
	// fmt.Println(string(input[0:o]))

	// Unmarshal answer to get it

	// Offer
	// offer := Offer{Opcode: "offer", Mac: , Payload: "Hallo Welt"}
}
