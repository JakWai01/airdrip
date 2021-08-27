package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net"
	"strings"
)

// This signaling protocol is heavily inspired by the weron project created by @pojntfx
// Take a look at the specification by clicking the following link: https://github.com/pojntfx/weron/blob/main/docs/signaling-protocol.txt#L12

var communities = map[string][]string{}
var macs = map[string]bool{}
var connections = map[string]net.Conn{}

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

type Rejection struct {
	Opcode string `json:"opcode"`
}

type Ready struct {
	Opcode string `json:"opcode"`
	Mac    string `json:"mac"`
}

type Introduction struct {
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

type Candidate struct {
	Opcode  string `json:"opcode"`
	Mac     string `json:"mac"`
	Payload string `json:"payload"`
}

type Exited struct {
	Opcode string `json:"opcode"`
}

type Resignation struct {
	Opcode string `json:"opcode"`
	Mac    string `json:"mac"`
}

func handleConnection(c net.Conn) {
	for {
		// error here
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			panic(err)
		}

		fmt.Println(message)

		values := make(map[string]json.RawMessage)

		err = json.Unmarshal([]byte(message), &values)
		if err != nil {
			panic(err)
		}

		switch Opcode(strings.ReplaceAll(string(values["opcode"]), "\"", "")) {
		case application:
			// we get community and mac. Check if community exists. If not create it. Only allow unused macs.

			// Community maps string to tuple. Macs is an array and must be unique.
			fmt.Println("application")
			var opcode Application

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			if _, ok := macs[opcode.Mac]; ok {
				// send Rejection. That Mac is already contained
				byteArray, err := json.Marshal(Rejection{Opcode: string(rejection)})
				if err != nil {
					panic(err)
				}

				fmt.Println(string(byteArray))

				_, err = c.Write(byteArray)
				if err != nil {
					panic(err)
				}
				break
			}

			// Store connection in a map
			connections[opcode.Mac] = c
			fmt.Println(connections)

			// check if community exists and if there are less than 2 members inside
			if val, ok := communities[opcode.Community]; ok {
				// check if length smaller than 2
				if len(val) >= 2 {
					// send Rejection. This community is full
					byteArray, err := json.Marshal(Rejection{Opcode: string(rejection)})
					if err != nil {
						panic(err)
					}

					fmt.Println(string(byteArray))

					_, err = c.Write(byteArray)
					if err != nil {
						panic(err)
					}
					break
				} else {
					// Community exists but has less than 2 values in it
					communities[opcode.Community] = append(communities[opcode.Community], opcode.Mac)
					fmt.Println(communities)

					macs[opcode.Mac] = false
					fmt.Println(macs)

					// send Acceptance
					byteArray, err := json.Marshal(Acceptance{Opcode: string(acceptance)})
					if err != nil {
						panic(err)
					}

					fmt.Println(string(byteArray))

					_, err = c.Write(byteArray)
					if err != nil {
						panic(err)
					}
					break
				}
			} else {
				// Community does not exist. Create community and insert mac
				communities[opcode.Community] = append(communities[opcode.Community], opcode.Mac)
				fmt.Println(communities)

				macs[opcode.Mac] = false
				fmt.Println(macs)

				// send Acceptance
				byteArray, err := json.Marshal(Acceptance{Opcode: string(acceptance)})
				if err != nil {
					panic(err)
				}

				fmt.Println(string(byteArray))

				_, err = c.Write(byteArray)
				if err != nil {
					panic(err)
				}
				break
			}

		case ready:
			var opcode Ready

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			// If we receive ready, mark the sending person as ready and check if both are ready. Loop through all communities to get the community the person is in.
			macs[opcode.Mac] = true

			// Loop through all members of the community and through all elements in it. If the mac isn't member of a community, this will panic.
			community, err := getCommunity(opcode.Mac)
			if err != nil {
				panic(err)
			}

			if len(communities[community]) == 2 {
				if macs[communities[community][0]] == true && macs[communities[community][1]] == true {
					// Send an introduction to the peer containing the address of the first peer.
					byteArray, err := json.Marshal(Introduction{Opcode: string(introduction), Mac: communities[community][0]})
					if err != nil {
						panic(err)
					}

					fmt.Println(string(byteArray))

					_, err = c.Write(byteArray)
					if err != nil {
						panic(err)
					}
					break

				}
			}

			break

		case offer:
			// Contains the mac of the receiver and a payload this receiver should receive
			fmt.Println("offer")
			var opcode Offer

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			// Get connection of the reveiver and send him the payload
			receiver := connections[opcode.Mac]

			var senderMac string
			// Get the Mac based on the current connection out of the connections mac
			for key, val := range connections {
				if c == val {
					senderMac = key
				}
			}

			byteArray, err := json.Marshal(Offer{Opcode: string(offer), Mac: senderMac, Payload: opcode.Payload})
			if err != nil {
				panic(err)
			}

			fmt.Println(string(byteArray))

			_, err = receiver.Write(byteArray)
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

			// Get connection of the receiver and send him the payload
			receiver := connections[opcode.Mac]

			var senderMac string
			// Get the Mac based on the current connection out of the connections mac
			for key, val := range connections {
				if c == val {
					senderMac = key
				}
			}

			byteArray, err := json.Marshal(Answer{Opcode: string(answer), Mac: senderMac, Payload: opcode.Payload})
			if err != nil {
				panic(err)
			}

			fmt.Println(string(byteArray))

			_, err = receiver.Write(byteArray)
			if err != nil {
				panic(err)
			}
		case candidate:
			fmt.Println("candidate")
			var opcode Candidate

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			// Get connection of the receiver and send him the payload
			receiver := connections[opcode.Mac]

			var senderMac string
			// Get the Mac based on the current connection out of the connections mac
			for key, val := range connections {
				if c == val {
					senderMac = key
				}
			}

			byteArray, err := json.Marshal(Candidate{Opcode: string(candidate), Mac: senderMac, Payload: opcode.Payload})
			if err != nil {
				panic(err)
			}

			fmt.Println(string(byteArray))

			_, err = receiver.Write(byteArray)
			if err != nil {
				panic(err)
			}
		case exited:
			fmt.Println("exited")
			var opcode Exited

			err := json.Unmarshal([]byte(message), &opcode)
			if err != nil {
				panic(err)
			}

			var senderMac string
			// Get the Mac based on the current connection out of the connections mac
			for key, val := range connections {
				if c == val {
					senderMac = key
				}
			}

			byteArray, err := json.Marshal(Resignation{Opcode: string(resignation), Mac: senderMac})
			if err != nil {
				panic(err)
			}

			fmt.Println(string(byteArray))

			var receiver net.Conn

			// Get the other peer in the community
			community, err := getCommunity(senderMac)
			if err != nil {
				panic(err)
			}

			if senderMac == communities[community][0] {
				// the second one is receiver
				receiver = connections[communities[community][1]]
			} else {
				// first one
				receiver = connections[communities[community][0]]
			}

			// Send to the other peer
			_, err = receiver.Write(byteArray)
			if err != nil {
				panic(err)
			}

			// Remove this peer from all maps
			delete(macs, senderMac)
			delete(connections, senderMac)

			// Remove community
			delete(communities, community)
		default:
			panic("Invalid message. Please use a valid opcode.")
		}
	}
}

func main() {
	var laddr = flag.String("laddr", "localhost:8080", "listen address")
	flag.Parse()

	fmt.Println(*laddr)

	l, err := net.Listen("tcp4", *laddr)
	if err != nil {
		panic(err)
	}

	// defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go handleConnection(c)
	}
}

func getCommunity(mac string) (string, error) {
	for key, element := range communities {
		for i := 0; i < len(element); i++ {
			if element[i] == mac {
				return key, nil
			}
		}
	}
	return "", errors.New("This mac is not part of any community so far!")
}
