package mdns

import (
	"context"
	"log"
	"net"

	"github.com/pion/mdns"
	"golang.org/x/net/ipv4"
)

func LookupMDNS(ch chan string) {
	addr, err := net.ResolveUDPAddr("udp", mdns.DefaultAddress)
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatal(err)
	}

	server, err := mdns.Server(ipv4.NewPacketConn(l), &mdns.Config{})
	if err != nil {
		log.Fatal(err)
	}

	_, src, err := server.Query(context.TODO(), "_signaling.local")

	ch <- src.String()
}
