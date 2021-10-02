ADDRESS := $(shell . run.sh)

server:
	go run main.go signal --address=${ADDRESS}
client:
	go run main.go client --mac=${MAC} --community=${COMMUNITY} --laddr=${LADDR}

mdns:
	go run main.go mdns

lookup:
	go run main.go lookup