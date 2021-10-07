package mdns

import (
	"log"
	"net"

	"github.com/pion/mdns"
	"golang.org/x/net/ipv4"
)

func RunMDNS() {
	addr, err := net.ResolveUDPAddr("udp", mdns.DefaultAddress)
	if err != nil {
		log.Fatal(err)
	}

	l, err := net.ListenUDP("udp4", addr)
	if err != nil {
		log.Fatal(err)
	}

	_, err = mdns.Server(ipv4.NewPacketConn(l), &mdns.Config{
		LocalNames: []string{"_signaling.local"},
	})
	if err != nil {
		log.Fatal(err)
	}
	select {}
}
