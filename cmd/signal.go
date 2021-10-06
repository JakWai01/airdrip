package cmd

import (
	"log"
	"net"
	"net/http"

	"github.com/JakWai01/airdrip/pkg/mdns"
	"github.com/JakWai01/airdrip/pkg/signaling"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"nhooyr.io/websocket"
)

const (
	addressKey = "address"
	portKey    = "port"
)

var signalCmd = &cobra.Command{
	Use:   "signal",
	Short: "Start a signaling server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle lifecycle
		signaler := signaling.NewSignalingServer()

		go func() {
			go mdns.RunMDNS()
		}()

		for {
			socket := viper.GetString(addressKey) + ":" + viper.GetString(portKey)

			addr, err := net.ResolveTCPAddr("tcp", socket)
			if err != nil {
				panic(err)
			}

			log.Printf("signaling server listening on %v", socket)

			handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				conn, err := websocket.Accept(rw, r, nil)
				if err != nil {
					panic(err)
				}

				log.Println("client connected")

				go func() {
					signaler.HandleConn(*conn)
				}()
			})

			http.ListenAndServe(addr.String(), handler)

		}
	},
}

func init() {
	signalCmd.PersistentFlags().String(addressKey, "localhost", "Listen address")
	signalCmd.PersistentFlags().String(portKey, "8080", "Port to listen on")

	// Bind env variables
	if err := viper.BindPFlags(signalCmd.PersistentFlags()); err != nil {
		log.Fatal("could not bind flags:", err)
	}
	viper.SetEnvPrefix("airdrip")
	viper.AutomaticEnv()
}
