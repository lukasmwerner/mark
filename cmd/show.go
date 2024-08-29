/*
Copyright © 2024 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
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
		db, err := store.Open()
		if err != nil {
			log.Panicln(err.Error())
			return
		}
		defer db.Close()

		searchQuery := strings.Join(args, " ")

		bookmarks, err := store.SearchBookmarks(db, searchQuery)
		if err != nil {
			log.Panicln("unable to search bookmarks", err.Error())
		}
		if len(bookmarks) == 0 {
			fmt.Println("found no bookmarks")
			return
		}

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
				log.Fatalln(err.Error())
			}
			bookmarks = []store.Bookmark{bookmarks[pickedIndex]}
		}

		outputBookmark(outputMode, bookmarks[0])
		return
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
	showCmd.Flags().StringVarP(&outputMode, "mode", "m", "json", "Output mode: json,csv")
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

	}
}
