/*
Copyright © 2025 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/x/term"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// Style definitions
var (
	promptStyle = lipgloss.NewStyle().
			Foreground(highlight).
			Bold(true)

	resultHeaderStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(highlight).
				Padding(0, 1)

	resultRowStyle = lipgloss.NewStyle().
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)
)

// sqlCmd represents the sql command
var sqlCmd = &cobra.Command{
	Use:   "sql",
	Short: "Lets you run sql queries on your database",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open()
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		defer db.Close()

		if len(args) > 0 {
			query := strings.Join(args, " ")
			executeQuery(db, query)
			return
		}
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("mark SQL REPL - Enter SQL queries to execute (type '.exit' or '.quit' to exit)")
		fmt.Println("Bookmarks are in the 'Bookmarks' table")
		fmt.Println("Type '.help' for available commands")

		var query string
		for {
			fmt.Print(promptStyle.Render("mark-sql> "))
			input, _ := reader.ReadString('\n')
			input = strings.TrimSpace(input)

			if input == "" {
				continue
			}

			if input == ".exit" || input == ".quit" || input == ".q" {
				break
			}

			if input == ".help" || input == "?" {
				printHelp()
				continue
			}

			if input == ".tables" {
				query = "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name;"
			} else if input == ".schema" {
				query = "SELECT sql FROM sqlite_master WHERE type='table' ORDER BY name;"
			} else if search, ok := strings.CutPrefix(input, ".search "); ok {
				rows, _ := db.Query("SELECT * FROM Bookmarks_fts WHERE Bookmarks_fts MATCH ?;", search)
				renderResults(rows)
				query = ""
				continue
			} else {
				// Append to current query if it doesn't end with semicolon
				if !strings.HasSuffix(query, ";") && query != "" {
					query += " " + input
				} else {
					query = input
				}
			}

			// Execute only if query ends with semicolon
			if strings.HasSuffix(query, ";") {
				rows, _ := executeQuery(db, query)
				renderResults(rows)
				query = ""
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(sqlCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sqlCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sqlCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func renderResults(rows *sql.Rows) {
	if rows == nil {
		return
	}
	columns, err := rows.Columns()
	if err != nil {
		fmt.Println(errorStyle.Render("Error getting columns: " + err.Error()))
		return
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	width, _, err := term.GetSize(0)
	if err != nil {
		fmt.Println(errorStyle.Render("Error getting terminal size: " + err.Error()))
		return
	}

	t := table.New().Headers(columns...).Width(width)

	rowCount := 0
	for rows.Next() {
		rowCount++
		err = rows.Scan(valuePtrs...)
		if err != nil {
			fmt.Println(errorStyle.Render("Error scanning row: " + err.Error()))
			continue
		}

		row := []string{}
		for _, val := range values {
			var displayVal string
			switch v := val.(type) {
			case []byte:
				displayVal = string(v)
			case nil:
				displayVal = "NULL"
			default:
				displayVal = fmt.Sprintf("%v", v)
			}
			row = append(row, resultRowStyle.Render(displayVal))
		}
		t.Row(row...)
	}

	fmt.Println(t.Render())

	if err = rows.Err(); err != nil {
		fmt.Println(errorStyle.Render("Error during iteration: " + err.Error()))
		return
	}

	fmt.Printf("Query executed successfully. %d row(s) returned.\n", rowCount)

}

func executeQuery(db *store.DB, query string) (*sql.Rows, error) {
	query = strings.TrimSpace(query)

	rows, err := db.Query(query)
	if err != nil {
		fmt.Println(errorStyle.Render("Error: " + err.Error()))
		return nil, err
	}
	return rows, err
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  .help, ?     - Show this help")
	fmt.Println("  .exit, .quit - Exit the REPL")
	fmt.Println("  .tables      - List all tables")
	fmt.Println("  .schema      - Show schema for all tables")
	fmt.Println("  .search 		- Use FTS search for tables")
	fmt.Println("")
	fmt.Println("Notes:")
	fmt.Println("- SQL queries can span multiple lines until a semicolon is entered")
	fmt.Println("- Press Enter on an empty line to submit the query")
}
