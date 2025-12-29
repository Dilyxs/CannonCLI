/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dilyxs/CannonCLI/internal/tui"
	"github.com/dilyxs/CannonCLI/pkg"
)

func main() {
	// cmd.Execute()
	r := pkg.RequestDetails{"GET", "http://localhost:8080", "", 2 * time.Second}
	CancelChan := make(chan bool, 100)
	OuputChan := make(chan pkg.ResponseWithStatus, 1)
	m := tui.InitModel(
		1000,
		5,
		1000,
		r,
		CancelChan,
		OuputChan,
	)
	p := tea.NewProgram(m)
	if finalstage, err := p.Run(); err != nil {
		fmt.Printf("ran into err:%v", err)
	} else {
		fmt.Printf("Final Model: %v", finalstage.View())
	}
}
