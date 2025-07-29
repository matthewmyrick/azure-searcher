package main

import (
	"log"

	"azure-searcher/src/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize the UI model
	model, err := ui.NewModel()
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Check Azure CLI prerequisites
	if err := model.AzureClient.CheckCLI(); err != nil {
		log.Fatal(err)
	}

	if err := model.AzureClient.CheckLogin(); err != nil {
		log.Fatal(err)
	}

	// Create and start the Bubble Tea program
	program := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := program.Run(); err != nil {
		log.Fatal(err)
	}
}