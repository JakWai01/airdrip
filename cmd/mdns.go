package cmd

import (
	"os"

	"github.com/JakWai01/airdrip/pkg/mdns"
	"github.com/spf13/cobra"
)

var mdnsCmd = &cobra.Command{
	Use:   "mdns",
	Short: "Start the mdns service.",
	RunE: func(cmd *cobra.Command, args []string) error {

		fatal := make(chan error)
		done := make(chan struct{})

		go func() {
			go mdns.RunMDNS()
		}()

		for {
			select {
			case err := <-fatal:
				panic(err)
			case <-done:
				os.Exit(0)
			}
		}
	},
}
