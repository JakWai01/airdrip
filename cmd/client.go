package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	laddrKey     = "laddr"
	communityKey = "community"
	macKey       = "mac"
)

var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "Start a signaling client.",
	RunE: func(cmd *cobra.Command, args []string) error {

		fatal := make(chan error)
		done := make(chan struct{})

		// client := signaling.NewSignalingClient()

		go func() {

			// go client.HandleConn(viper.GetString(laddrKey), viper.GetString(communityKey), viper.GetString(macKey))

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

func init() {
	clientCmd.PersistentFlags().String(laddrKey, "localhost:8080", "Listen address")
	clientCmd.PersistentFlags().String(communityKey, "a", "Community to join")
	clientCmd.PersistentFlags().String(macKey, "124", "Mac to identify you as a unique host")

	// Bind env variables
	if err := viper.BindPFlags(clientCmd.PersistentFlags()); err != nil {
		log.Fatal("could not bind flags:", err)
	}
	viper.SetEnvPrefix("airdrip")
	viper.AutomaticEnv()
}
