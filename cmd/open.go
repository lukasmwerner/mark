/*
Copyright © 2024 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/cli/browser"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// openCmd represents the open command
var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Opens the matching search link in the user's browser",
	Long:  ``,
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open(store.Options{
			Flags: []store.Flag{store.Embedding},
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer db.Close()

		searchQuery := strings.Join(args, " ")

		bookmarks, err := store.SearchBookmarks(db, searchQuery)
		if err != nil {
			fmt.Println("unable to search bookmarks", err.Error())
			return
		}
		if len(bookmarks) == 0 {
			fmt.Println("found no bookmarks")
			return
		}

		if len(bookmarks) == 1 {
			fmt.Printf("Opening %s %s\n", bookmarks[0].Title, bookmarks[0].Url)
			browser.OpenURL(bookmarks[0].Url)
			return
		}

		pickedLink := ""
		options := make([]huh.Option[string], len(bookmarks))
		for i, bookmark := range bookmarks {
			options[i] = huh.NewOption(bookmark.Title, bookmark.Url)
		}
		err = huh.NewSelect[string]().Title("Pick your link").Options(options...).Value(&pickedLink).Run()
		if err != nil {
			if err == huh.ErrUserAborted {
				return
			}
			fmt.Println(err.Error())
			return
		}

		err = browser.OpenURL(pickedLink)
		if err != nil {
			fmt.Println("unable to open url in browser", err.Error())
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(openCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// openCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// openCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
