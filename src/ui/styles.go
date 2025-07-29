package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	SelectedStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#874BFD")).
		Foreground(lipgloss.Color("#FFFFFF"))

	ResourceGroupStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#0078D4")). // Azure blue
		Bold(true)

	ResourceStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E0E0E0")) // Lighter gray for better visibility

	CacheStatusStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("244"))
)