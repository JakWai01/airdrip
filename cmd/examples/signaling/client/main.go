package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
)

type Application struct {
	Opcode    string `json:"opcode"`
	Community string `json:"community"`
	Mac       string `json:"mac"`
}

type Ready struct {
	Opcode string `json:"opcode"`
	Mac    string `json:"mac"`
}

// take flags for community and mac
func main() {
	var laddr = flag.String("laddr", "localhost:8080", "listen address")
	var mac = flag.String("mac", "123", "mac (identification string)")
	var community = flag.String("community", "a", "community to join")
	flag.Parse()

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

	var input [1024]byte

	o, err := conn.Read(input[0:])
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println(string(input[0:o]))

	ready := Ready{Opcode: "ready", Mac: *mac}

	byteArray, err = json.Marshal(ready)
	if err != nil {
		panic(err)
	}

	byteArray = append(byteArray, "\n"...)
	_, err = conn.Write([]byte(byteArray))
	if err != nil {
		panic(err)
	}

	o, err = conn.Read(input[0:])
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println(string(input[0:o]))

	select {}
}
