package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
)

// This signaling protocol is heavily inspired by the weron project created by @pojntfx
// Take a look at the specification by clicking the following link: https://github.com/pojntfx/weron/blob/main/docs/signaling-protocol.txt#L12

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
	opcode    string
	community string
	mac       string
}

type Acceptance struct{}

type Rejection struct{}

type Ready struct{}

type Introduction struct {
	mac string
}

type Offer struct {
	mac     string
	payload string
}

type Answer struct {
	mac     string
	payload string
}

type Candidate struct {
	mac     string
	payload string
}

type Exited struct{}

type Resignation struct {
	mac string
}

func handleConnection(c net.Conn) {
	for {
		message, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			panic(err)
		}

		// a map container to decode the JSON structure into
		values := make(map[string]json.RawMessage)

		// unmarshal JSON
		err = json.Unmarshal([]byte(message), &values)
		if err != nil {
			panic(err)
		}

		fmt.Printf("%v", values["hallo"])
		fmt.Println()

		result := message + "\n"
		c.Write([]byte(string(result)))
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

	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			panic(err)
		}
		go handleConnection(c)
	}
}
