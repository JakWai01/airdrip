package mdns

import (
	"fmt"

	"github.com/hashicorp/mdns"
)

func LookupMDNS(ch chan struct{}) {
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 4)
	go func() {
		for entry := range entriesCh {
			fmt.Printf("Got new entry: %v\n", entry)
		}
	}()

	// Start the lookup
	mdns.Lookup("_airdrip._tcp", entriesCh)
	close(entriesCh)
	ch <- struct{}{}
}
