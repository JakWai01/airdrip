package signal

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
)

// HTTPSDPServer starts a HTTP Server that consumes SDPs
func HTTPSDPServer() chan string {

	// parse "port" flag
	port := flag.Int("port", 8080, "http server port")
	flag.Parse()

	// create sdpchannel
	sdpChan := make(chan string)

	// create /sdp endpoint
	http.HandleFunc("/sdp", func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Fprintf(w, "done")
		sdpChan <- string(body)
	})

	// serve server
	go func() {
		err := http.ListenAndServe(":"+strconv.Itoa(*port), nil)
		if err != nil {
			panic(err)
		}
	}()

	// return the channel
	return sdpChan
}
