OS_NAME := $(shell uname -s | tr A-Z a-z)
ifeq ($(OS_NAME), darwin) 
	ADDRESS = $(shell . run.sh)
endif
ifeq ($(OS_NAME), linux) 
	ADDRESS = $(shell bash run.sh)
endif

server:
	go run main.go signal --address=${ADDRESS}
client:
	go run main.go client --community=${COMMUNITY} --laddr=${LADDR}

mdns:
	go run main.go mdns

lookup:
	go run main.go lookup