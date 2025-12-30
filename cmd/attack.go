/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dilyxs/CannonCLI/internal/tui"
	"github.com/dilyxs/CannonCLI/pkg"
	"github.com/spf13/cobra"
)

// attackCmd represents the attack command
var attackCmd = &cobra.Command{
	Use:   "attack",
	Short: "Lauch Request To OverLoad Server and see its capabilities",
	Long: `Attack Command will launch a predefined number of requests to the target server.
		User Defines paramters like: RequestPerSecond, TotalRequests, Timeout, TargetURL etc.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 1 {
			fmt.Printf("only single link permittable!: %v", args)
			return
		}
		link := strings.Join(args, "")
		method, err := cmd.Flags().GetString("method")
		if err != nil {
			fmt.Println("Error getting 'method':", err)
			return
		}
		bodyfile, err := cmd.Flags().GetString("body")
		if err != nil {
			fmt.Println("Error getting 'body':", err)
			return
		}
		rps, err := cmd.Flags().GetInt("RequestPerSecond")
		if err != nil {
			fmt.Println("Error getting 'RequestPerSecond':", err)
			return
		}
		TotalSeconds, err := cmd.Flags().GetInt("TotalSeconds")
		if err != nil {
			fmt.Println("Error getting 'TotalSeconds':", err)
			return
		}
		workerCount, err := cmd.Flags().GetInt("WorkerCount")
		if err != nil {
			fmt.Println("Error getting 'WorkerCount':", err)
			return
		}
		RequestToBeSent := pkg.RequestDetails{
			Method:    method,
			Link:      link,
			Filepath:  bodyfile,
			TimeLimit: pkg.TimeLimitForFetcher,
		}
		CancelChan := make(chan bool, 100)
		OuputChan := make(chan pkg.ResponseWithStatus, 1)
		m := tui.InitModel(
			int(rps),
			int(TotalSeconds),
			workerCount,
			RequestToBeSent,
			CancelChan,
			OuputChan,
		)
		p := tea.NewProgram(m)
		if _, err := p.Run(); err != nil {
			fmt.Printf("ran into err:%v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(attackCmd)

	// Here you will define your flags and configuration settings.
	attackCmd.Flags().IntP("RequestPerSecond", "p", 200, "how many requests per Second")
	attackCmd.Flags().IntP("TotalSeconds", "n", 5, "Time during which we will send requests")
	attackCmd.Flags().IntP("WorkerCount", "w", 200, "number of workers")

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// attackCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// attackCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
