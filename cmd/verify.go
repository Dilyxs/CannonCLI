/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dilyxs/CannonCLI/pkg"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify",
	Short: "See the Headers&&Body of a single request",
	Long:  `Sends a single request using the provided flags and prints the full raw response (Headers + Body)`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			fmt.Printf("only single link permittable!: %v", args)
			return
		}
		link := strings.Join(args, "")
		method, _ := cmd.PersistentFlags().GetString("method")
		bodyfile, _ := cmd.PersistentFlags().GetString("body")
		body := make(map[string]any)
		if bodyfile != "" {
			details, err := os.ReadFile(bodyfile)
			if err != nil {
				fmt.Printf("file does not exist!: %v or no body file provided!", bodyfile)
			}
			if err := json.Unmarshal(details, &body); err != nil {
				fmt.Printf("file contains weird json, cannot decode!: %v", bodyfile)
				return
			}
		}
		r, latency, err := pkg.Fetch(method, link, "application/json", 2*time.Second, body)
		if err != nil {
			fmt.Println("FAILED TO FETCH!!!")
			return
		}
		response := make(map[string]any)
		var IsOk bool
		if err := json.NewDecoder(r.Body).Decode(&response); err != nil && r.StatusCode == 200 {
			IsOk = true
		}

		message := fmt.Sprintf("target: %v\n method:%v\n body:%v\n Response Status: %v\n latency: %v\n Response:%v\n. verification:%v",
			link, method, body, r.StatusCode, latency, response, IsOk)
		fmt.Println(message)
	},
}

func init() {
	rootCmd.AddCommand(verifyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	verifyCmd.PersistentFlags().StringP("method", "m", "GET", "Define method for request")
	verifyCmd.PersistentFlags().StringP("body", "b", "", "Path to a local file (e.g., ./payload.json). Read this file once into memory at startup, then reuse the byte slice for every request. Do not read from disk on every loop.")
	// verifyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// verifyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
