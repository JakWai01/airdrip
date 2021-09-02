package cmd

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/JakWai01/airdrip/pkg/signaling"
	"github.com/spf13/cobra"
)

var signalCmd = &cobra.Command{
	Use:   "signal",
	Short: "Start a signaling server.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle lifecycle
		fatal := make(chan error)
		done := make(chan struct{})
		signaler := signaling.NewSignalingServer()

		go func() {

			l, err := net.Listen("tcp", "localhost:8080")
			if err != nil {
				panic(err)
			}

			fmt.Println("signaling server listening on localhost:8080")
			defer l.Accept()

			for {
				c, err := l.Accept()
				if err != nil {
					panic(err)
				}

				go signaler.HandleConn(c)
			}
		}()

		for {
			select {
			case err := <-fatal:
				log.Fatal(err)
			case <-done:
				os.Exit(0)
			}
		}
	},
}
