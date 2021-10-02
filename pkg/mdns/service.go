package mdns

import (
	"os"

	"github.com/hashicorp/mdns"
)

func RunMDNS() {
	host, _ := os.Hostname()
	info := []string{"Discovering hosts with MDNS for the airdrip project"}
	service, _ := mdns.NewMDNSService(host, "_airdrip._tcp", "", "", 8000, nil, info)

	// Create the mDNS server, defer shutdown
	server, _ := mdns.NewServer(&mdns.Config{Zone: service})
	defer server.Shutdown()

	select {}
}
