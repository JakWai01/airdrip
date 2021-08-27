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

func main() {
	var laddr = flag.String("laddr", "localhost:8080", "listen address")
	flag.Parse()

	conn, err := net.Dial("tcp", *laddr)
	if err != nil {
		panic(err)
	}

	application := Application{Opcode: "application", Community: "a", Mac: "123"}
	// _, err = conn.Write([]byte(`{"opcode":"application", "community":"a", "mac":"123"}` + "\n"))

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
}
