package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	var laddr = flag.String("laddr", "localhost:8080", "listen address")
	flag.Parse()

	conn, err := net.Dial("tcp4", *laddr)
	if err != nil {
		panic(err)
	}

	userInput := bufio.NewReader(os.Stdin)
	for {
		userLine, err := userInput.ReadBytes(byte('\n'))
		switch err {
		case nil:
			conn.Write(userLine)
		case io.EOF:
			os.Exit(0)
		default:
			panic(err)
		}

		var input [2048]byte

		o, err := conn.Read(input[0:])
		if err != nil {
			panic(err)
		}

		fmt.Println(string(input[0:o]))
	}
}
