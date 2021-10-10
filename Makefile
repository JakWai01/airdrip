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
	go run main.go client --community=${COMMUNITY} --laddr=${LADDR} --port=${PORT}
mdns:
	go run main.go mdns
lookup:
	go run main.go lookup
run: 
	GOOS=js GOARCH=wasm go build -o  cmd/examples/webassembly/assets/json.wasm cmd/examples/webassembly/cmd/wasm/main.go
	go run ./cmd/examples/webassembly/cmd/server/main.go