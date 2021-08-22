package main

import (
	"flag"
	"fmt"
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

func main() {
	var laddr = flag.String("laddr", "localhost:8080", "listen address")
	flag.Parse()

	fmt.Println(*laddr)
}
