/*
Copyright © 2025 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"
	"log"

	_ "embed"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/lukasmwerner/mark/store"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
)

//go:embed logo.ascii
var markascii string

var blueLink = lipgloss.NewStyle().Foreground(lipgloss.Color("#8FA8FF"))
var tagsStyling = lipgloss.NewStyle().Foreground(lipgloss.Color("#DBBC7F"))

type tuiAppModel struct {
	db        *store.DB
	searchBox textinput.Model
	viewport  viewport.Model
	bookmarks []store.Bookmark
	width     int
	height    int
	ready     bool
}

func (m tuiAppModel) Init() tea.Cmd { return nil }

func (m tuiAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-2)
			m.viewport.YPosition = 2
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - 2
		}
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			m.searchBox.Blur()
			bookmarks, err := store.SearchBookmarks(m.db, m.searchBox.Value())
			if err != nil {
				log.Fatalln("search failed")
			}
			m.bookmarks = bookmarks

			results := ""
			for _, bm := range m.bookmarks {
				results += fmt.Sprintf("%s <%s>\n%s", blueLink.Render(termenv.Hyperlink(bm.Url, bm.Title)), tagsStyling.Render(fmt.Sprint(bm.Tags)), bm.Description)
				results += "\n---\n"
			}
			m.viewport.SetContent(results)
		}
	}
	if m.searchBox.Focused() {
		var cmd tea.Cmd
		m.searchBox, cmd = m.searchBox.Update(msg)
		cmds = append(cmds, cmd)
	} else {
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}
func (m tuiAppModel) View() string {
	out := ""
	if m.searchBox.Focused() {
		searchBox := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Render(m.searchBox.View())
		out = lipgloss.JoinVertical(lipgloss.Center, markascii, searchBox)
		out = lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, out)
	} else {
		out := "mark: results for> " + m.searchBox.Value()
		out = lipgloss.NewStyle().
			Border(lipgloss.BlockBorder()).
			BorderBottom(true).
			BorderLeft(false).
			BorderRight(false).
			BorderTop(false).
			Width(m.width).
			Render(out)
		return out + "\n" + m.viewport.View()
	}

	return out

}

// searchCmd represents the search tui command
var searchCmd = &cobra.Command{
	Use:   "search",
	Short: `[EXPERIMENTAL] google search like tui intended for hosting via gotty or ttyd`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open()
		if err != nil {
			log.Fatalln("unable to connect to backing store")
		}
		input := textinput.New()
		input.Placeholder = "Search"
		input.Prompt = ""
		input.Focus()
		input.Width = 60

		m := tuiAppModel{
			db:        db,
			searchBox: input,
		}

		prog := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())

		if _, err := prog.Run(); err != nil {
			fmt.Println("Error running program:", err.Error())
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchTuiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchTuiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
