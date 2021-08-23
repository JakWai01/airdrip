package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"strings"
)

// This signaling protocol is heavily inspired by the weron project created by @pojntfx
// Take a look at the specification by clicking the following link: https://github.com/pojntfx/weron/blob/main/docs/signaling-protocol.txt#L12

type Application struct {
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

		temp := strings.TrimSpace(string(message))
		fmt.Println(temp)

		result := temp + "\n"
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
