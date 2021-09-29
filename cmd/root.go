package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "airdrip",
	Short: "Like Aidrop, but with more drip!",
	Long: `File sharing like Airdrop, but with more drip."

For more information, please visit https://github.com/JakWai01/airdrip`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	rootCmd.AddCommand(signalCmd)
	rootCmd.AddCommand(clientCmd)
}
