package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/JakWai01/airdrip/pkg/mdns"
	"github.com/spf13/cobra"
)

var lookupCmd = &cobra.Command{
	Use:   "lookup",
	Short: "Start the mdns lookup",
	RunE: func(cmd *cobra.Command, args []string) error {

		fatal := make(chan error)
		done := make(chan string)

		go func() {
			go mdns.LookupMDNS(done)
		}()

		for {
			select {
			case err := <-fatal:
				panic(err)
			case socket := <-done:
				ip := strings.Split(socket, ":")[0]
				fmt.Println(ip)

				os.Exit(0)
			}
		}
	},
}
