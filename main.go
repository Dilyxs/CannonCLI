/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"time"

	"github.com/dilyxs/CannonCLI/pkg"
)

func main() {
	// cmd.Execute()
	cancel := make(chan bool)
	r := pkg.RequestDetails{"GET", "http://localhost:8080", "", 2 * time.Second}
	pkg.RunAction(20, 4, r, 20, cancel)
}
