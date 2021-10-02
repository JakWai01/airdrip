package mdns

import (
	"context"
	"fmt"
	"net"

	"github.com/pion/mdns"
	"golang.org/x/net/ipv4"
)

func LookupMDNS(ch chan struct{}) {
	addr, err := net.ResolveUDPAddr("udp", mdns.DefaultAddress)
	if err != nil {
		panic(err)
	}

	l, err := net.ListenUDP("udp4", addr)
	if err != nil {
		panic(err)
	}

	server, err := mdns.Server(ipv4.NewPacketConn(l), &mdns.Config{})
	if err != nil {
		panic(err)
	}

	answer, src, err := server.Query(context.TODO(), "_airdrip.local")
	fmt.Println(src)
	fmt.Println(answer)
	ch <- struct{}{}
}
