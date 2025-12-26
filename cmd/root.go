/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "CannonCLI",
	Short: "Tool to Load Test an API Endpoint!",
	Long:  `CLI tool where you run cannon "attack https://my-api.com --rate 5000" and it attempts to melt the server, providing live latency statistics (P99, P50)`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("no-color", "n", false, "Disable TUI colors")
	rootCmd.Flags().BoolP("verbose", "v", false, "Enables debug logs (e.g., printing specific connection errors)")
}
