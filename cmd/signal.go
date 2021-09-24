package cmd

import (
	"log"
	"net"
	"net/http"

	"github.com/JakWai01/airdrip/pkg/signaling"
	"github.com/spf13/cobra"
	"nhooyr.io/websocket"
)

var signalCmd = &cobra.Command{
	Use:   "signal",
	Short: "Start a signaling server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle lifecycle
		signaler := signaling.NewSignalingServer()

		for {
			laddr := "localhost:8080"

			addr, err := net.ResolveTCPAddr("tcp", laddr)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("signaling server listening on %v", laddr)

			handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				conn, err := websocket.Accept(rw, r, nil)
				if err != nil {
					log.Fatal(err)
				}

				log.Println("client connected")

				go func() {
					signaler.HandleConn(conn)
				}()
			})

			http.ListenAndServe(addr.String(), handler)
		}
	},
}
