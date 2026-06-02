/*
Copyright © 2024 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lukasmwerner/mark/search"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

var outputMode string

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Shows the entry of a bookmark",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open(store.Options{
			Flags: []store.Flag{store.Embedding},
		})
		if err != nil {
			fmt.Println("unable to open database", err.Error())
			return
		}
		defer db.Close()

		searchQuery := strings.Join(args, " ")

		results, err := search.FullText5(db, searchQuery)
		if err != nil {
			fmt.Println("unable to search bookmarks", err.Error())
			return
		}
		if len(results) == 0 {
			fmt.Println("found no bookmarks")
			return
		}
		bookmarks := search.ToBookmarks(results)

		if len(bookmarks) != 1 {
			pickedIndex := 0
			options := make([]huh.Option[int], len(bookmarks))
			for i, bookmark := range bookmarks {
				options[i] = huh.NewOption(bookmark.Title, i)
			}
			err = huh.NewSelect[int]().Title("Pick your link").Options(options...).Value(&pickedIndex).Run()
			if err != nil {
				if err == huh.ErrUserAborted {
					return
				}
				fmt.Println("unable to pick bookmark", err.Error())
				return
			}
			bookmarks = []store.Bookmark{bookmarks[pickedIndex]}
		}

		outputBookmark(outputMode, bookmarks[0])
	},
}

func init() {
	rootCmd.AddCommand(showCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// showCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// showCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	showCmd.Flags().StringVarP(&outputMode, "mode", "m", "pretty", "Output mode: json,csv,pretty")
}

func outputBookmark(mode string, bookmark store.Bookmark) {
	switch mode {
	case "json":
		b, _ := json.Marshal(bookmark)
		os.Stdout.Write(b)
	case "csv":
		w := csv.NewWriter(os.Stdout)
		w.Write([]string{"Title", "Description", "Tags", "URL"})
		w.Write([]string{bookmark.Title, bookmark.Description, strings.Join(bookmark.Tags, ","), bookmark.Url})
		w.Flush()
	case "pretty":
		fmt.Print(prettyOutput(bookmark))
	}
}

var categoryColorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#34A77C"))

func prettyOutput(bm store.Bookmark) string {
	out := ""
	out += categoryColorStyle.Render("Title: ") + bm.Title + "\n"
	out += categoryColorStyle.Render("URL: ") + bm.Url + "\n"
	out += categoryColorStyle.Render("Tags: ") + strings.Join(bm.Tags, ", ") + "\n"
	out += categoryColorStyle.Render("Description: ") + bm.Description + "\n"
	return out
}
