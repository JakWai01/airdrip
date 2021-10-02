server: 
	go run main.go signal

client:
	go run main.go client --mac=${MAC} --community=${COMMUNITY}

mdns:
	go run main.go mdns