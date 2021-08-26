package main

import (
	"flag"
	"fmt"
	"net"
)

func main() {
	var laddr = flag.String("laddr", "localhost:8080", "listen address")
	flag.Parse()

	conn, err := net.Dial("tcp4", *laddr)
	if err != nil {
		panic(err)
	}

	for {
		conn.Write([]byte(`{"opcode":"application", "community":"a", "mac":"123"}`))

		var input [2048]byte

		o, err := conn.Read(input[0:])
		if err != nil {
			panic(err)
		}

		fmt.Println(string(input[0:o]))
	}
}
