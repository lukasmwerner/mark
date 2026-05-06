/*
Copyright © 2024 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tables "github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/cli/browser"
	"github.com/lukasmwerner/mark/helpers"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

type mode string

var (
	NORMAL  mode = "NORMAL"
	SEARCH  mode = "SEARCH"
	PREVIEW mode = "PREVIEW"
)

var (
	normalModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("39"))

	searchModeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("251"))

	previewModeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("114"))

	statusBackground = lipgloss.Color("238")

	tableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder())

	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(false)

	unactiveSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("229")).
				Background(lipgloss.Color("#d2b2ff")).
				Bold(false)

	modalstyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder())
)

type rootAppModel struct {
	db           *store.DB
	input        textinput.Model
	table        tables.Model
	currentIndex int
	width        int
	height       int
	mode         mode
	rowsCount    int
}

func (m rootAppModel) Init() tea.Cmd { return nil }

func (m rootAppModel) updateTable() rootAppModel {

	bmCount := m.rowsCount
	var bookmarks []store.Bookmark
	var err error
	if m.input.Value() != "" {
		bookmarks, err = store.SearchBookmarks(m.db, m.input.Value())
		if err != nil {
			log.Panicln(err)
			return m
		}
	} else {
		bookmarks, err = store.GetBookmarks(m.db)
		if err != nil {
			log.Panicln(err)
			return m
		}
	}

	if len(bookmarks) == bmCount {
		return m
	}

	m.rowsCount = len(bookmarks)
	rows := make([]tables.Row, m.rowsCount)

	for i, bookmark := range bookmarks {
		rows[i] = []string{strings.TrimSpace(bookmark.Title), bookmark.Description, strings.Join(bookmark.Tags, ","), bookmark.Url}
	}
	m.table.SetRows(rows)

	return m
}

func (m rootAppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width
		m.table.SetColumns(resizeCols(msg.Width-10, msg.Height, m.table.Columns()))
		m.table.SetHeight(msg.Height - 4)
		m.table.SetWidth(msg.Width - 2)
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			if m.mode == NORMAL {
				return m, tea.Quit
			}
		case "esc":
			switch m.mode {
			case SEARCH:
				m.input.Blur()
				m.table.Focus()
				m = m.updateTable()
			case PREVIEW:
				m.table.Focus()
			}
			m.mode = NORMAL

		case " ":
			switch m.mode {
			case NORMAL:
				m.mode = PREVIEW
				m.table.Blur()
			case PREVIEW:
				m.mode = NORMAL
				m.table.Focus()
			}

		case "enter":
			switch m.mode {
			case NORMAL:
				url := m.table.SelectedRow()[3]
				browser.OpenURL(url)
			case PREVIEW:
				url := m.table.SelectedRow()[3]
				browser.OpenURL(url)
				m.mode = NORMAL
			case SEARCH:
				m.mode = NORMAL
				m.input.Blur()
				m.table.SetCursor(0)
				m = m.updateTable()
			}
		case "i":
			if m.mode == NORMAL {
				m.mode = SEARCH
				cmds = append(cmds, m.input.Focus())
				return m, tea.Batch(cmds...)
			}
		}
	}

	if m.mode == SEARCH && m.input.Focused() {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)
	}

	if m.mode == NORMAL {
		var cmd tea.Cmd
		m.table, cmd = m.table.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func resizeCols(width, _ int, cols []tables.Column) []tables.Column {
	newCols := make([]tables.Column, 4)
	copy(newCols, cols)

	newCols[0].Width = int(width / 4) // Title
	newCols[1].Width = int(width / 4) // Description
	newCols[2].Width = int(width / 4) // Tags
	newCols[3].Width = int(width / 4) // URL
	return newCols
}

func (m rootAppModel) View() string {
	statusBar := ""
	switch m.mode {
	case NORMAL:
		statusBar = normalModeStyle.Render(" " + string(m.mode) + " ")
	case SEARCH:
		statusBar = searchModeStyle.Render(" " + string(m.mode) + " ")
	case PREVIEW:
		statusBar = previewModeStyle.Render(" " + string(m.mode) + " ")
	}
	statusBar = lipgloss.PlaceHorizontal(m.width, lipgloss.Left, statusBar, lipgloss.WithWhitespaceBackground(statusBackground))

	table := lipgloss.PlaceVertical(m.height-2, lipgloss.Top, tableStyle.Render(m.table.View()))

	screen := lipgloss.JoinVertical(lipgloss.Left, m.input.View(), table, statusBar)

	if m.mode == PREVIEW {

		textSty := lipgloss.NewStyle().Width(m.width - 4).Align(lipgloss.Left)

		contents := categoryColorStyle.Render("title: ") + m.table.SelectedRow()[0] + "\n"
		contents += categoryColorStyle.Render("tags: ") + m.table.SelectedRow()[2] + "\n"
		contents += categoryColorStyle.Render("url: ") + m.table.SelectedRow()[3] + "\n"
		contents += categoryColorStyle.Render("desc: ") + textSty.Render(m.table.SelectedRow()[1])

		modal := modalstyle.Width(2 * (m.width / 3)).Render(contents)
		modalLeft := (m.width / 2) - (lipgloss.Width(modal) / 2)
		modalRight := (m.height / 2) - (lipgloss.Height(modal) / 2)
		return helpers.PlaceOverlay(modalLeft, modalRight, modal, screen, true)
	}

	return screen
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

		db, err := store.Open(store.Options{
			Flags: []store.Flag{store.Embedding},
		})
		if err != nil {
			fmt.Println("unable to open database", err.Error())
			return
		}

		defer db.Close()

		t := tables.New(tables.WithColumns([]tables.Column{
			{
				Title: "Title",
				Width: 10,
			},
			{
				Title: "Description",
				Width: 10,
			},
			{
				Title: "Tags",
				Width: 10,
			},
			{
				Title: "URL",
				Width: 10,
			},
		}))
		tableStyles := tables.DefaultStyles()
		tableStyles.Selected = selectedStyle
		t.SetStyles(tableStyles)
		t.Focus()
		t.KeyMap.PageDown = key.NewBinding(key.WithKeys("f", "pgdown"), key.WithHelp("f/pgdn", "page down"))

		input := textinput.New()
		input.Placeholder = "Search / Filter"

		m := rootAppModel{db: db, table: t, input: input, currentIndex: 1, rowsCount: 0, mode: NORMAL}
		m = m.updateTable()

		prog := tea.NewProgram(m, tea.WithAltScreen())

		if _, err := prog.Run(); err != nil {
			fmt.Println("Error running program:", err.Error())
			return
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
