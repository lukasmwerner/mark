package cmd

import "github.com/charmbracelet/lipgloss"


var (

	highlight = lipgloss.AdaptiveColor{Light: "#00A97A", Dark: "#6FC28E"}

	borders = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(highlight).
			Padding(1, 0)


	baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(highlight)
)


