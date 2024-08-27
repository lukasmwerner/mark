/*
Copyright © 2024 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/cli/browser"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

type mode string

var (
	NORMAL mode = "NORMAL"
	SEARCH mode = "SEARCH"
)

var (
	normalModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("39"))

	searchModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("251"))

	statusBackground = lipgloss.Color("238")

	headerStyle = lipgloss.NewStyle().
			BorderForeground(lipgloss.Color("240")).
			Bold(true)

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)

	unactiveSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("#d2b2ff")).
				Bold(false)
)

type rootAppModel struct {
	db           *store.DB
	input        textinput.Model
	table        *table.Table
	rows         []store.Bookmark
	currentIndex int
	width        int
	height       int
	mode         mode
	rowsCount    int
}

func (m rootAppModel) Init() tea.Cmd { return nil }

func (m rootAppModel) updateTable() rootAppModel {

	bookmarks, err := store.SearchBookmarks(m.db, m.input.Value())
	if err != nil {
		log.Panicln(err)
		return m
	}

	if m.rowsCount != 0 {
		m.table.ClearRows()
		m.table.Data(table.NewStringData())
	}

	m.rows = bookmarks
	m.rowsCount = len(bookmarks)
	m.currentIndex = 1

	for _, bookmark := range bookmarks {
		m.table.Row(strings.TrimSpace(bookmark.Title), bookmark.Description, strings.Join(bookmark.Tags, ","), bookmark.Url)
	}

	return m
}

func (m rootAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.table.Width(msg.Width).Height(msg.Height - 3)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.mode == NORMAL {
				return m, tea.Quit
			}
		case "esc":
			m.mode = NORMAL
			m.input.Blur()

			m = m.updateTable()
			break
		case "enter":
			switch m.mode {
			case NORMAL:
				// todo open active link
				if m.currentIndex <= m.rowsCount {
					url := m.rows[m.currentIndex-1].Url
					browser.OpenURL(url)
				}
			case SEARCH:
				m.mode = NORMAL
				m.input.Blur()
				m = m.updateTable()
				break
			}
		case "i":
			if m.mode == NORMAL {
				m.mode = SEARCH
				cmds = append(cmds, m.input.Focus())
				return m, tea.Batch(cmds...)
			}
		case "j", "up":
			if m.mode == NORMAL && m.currentIndex+1 <= m.rowsCount {
				m.currentIndex += 1
			}
		case "k", "down":
			if m.mode == NORMAL && m.currentIndex-1 != 0 {
				m.currentIndex -= 1
			}
		}

	}

	if m.mode == SEARCH && m.input.Focused() {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m rootAppModel) View() string {

	m.table.StyleFunc(func(row, col int) lipgloss.Style {
		switch {
		case row == 0:
			return headerStyle
		case row == m.currentIndex && m.mode == NORMAL:
			return selectedStyle
		case row == m.currentIndex && m.mode == SEARCH:
			return unactiveSelectedStyle
		default:
			return lipgloss.NewStyle()
		}
	})

	statusBar := ""
	switch m.mode {
	case NORMAL:
		statusBar = normalModeStyle.Render(" " + string(m.mode) + " ")
	case SEARCH:
		statusBar = searchModeStyle.Render(" " + string(m.mode) + " ")
	}

	statusBar = lipgloss.PlaceHorizontal(m.width, lipgloss.Left, statusBar, lipgloss.WithWhitespaceBackground(statusBackground))

	table := lipgloss.PlaceVertical(m.height-2, lipgloss.Top, m.table.Render())

	return lipgloss.JoinVertical(lipgloss.Left, m.input.View(), table, statusBar)
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "mark",
	Short: "A simple bookmark manager from the commandline",
	Long: `Mark is a simple bookmark manager that allows you to save and recall bookmarks.
It also allows you to sync those changes across all your devices using a 
file sync service. This is sort-of explained the following blog post: 
	https://lukaswerner.com/post/2024-08-13@Sqlite-Local-First`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		db, err := store.Open()
		if err != nil {
			log.Panicln(err)
			return
		}

		defer db.Close()

		t := table.New()

		input := textinput.New()
		input.Placeholder = "Search / Filter"

		m := rootAppModel{db: db, table: t, input: input, currentIndex: 1, rowsCount: 0, mode: NORMAL}

		t.Border(lipgloss.NormalBorder())

		t.Headers("Title", "Description", "Tags", "URL")

		prog := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := prog.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.mark.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
