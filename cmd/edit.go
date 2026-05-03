/*
Copyright © 2024 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// editCmd represents the edit command
var editCmd = &cobra.Command{
	Use:   "edit [search query]",
	Short: "Edit a bookmark",
	Long:  ``,
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
		if searchQuery == "" {
			err = huh.NewInput().Title("Search query").Value(&searchQuery).Run()
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			if searchQuery == "" {
				fmt.Println("no search query provided")
				return
			}
		}

		bookmarks, err := store.SearchBookmarks(db, searchQuery)
		if err != nil {
			fmt.Println("unable to search bookmarks", err.Error())
			return
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
				fmt.Println(err.Error())
				return
			}
			bookmarks = []store.Bookmark{bookmarks[pickedIndex]}
		}

		bookmark := bookmarks[0]
		original_bookmark := bookmark

		tags := strings.Join(bookmark.Tags, ",")

		huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Title").Value(&bookmark.Title),
			huh.NewInput().Title("Tags").Value(&tags),
			huh.NewInput().Title("URL").Value(&bookmark.Url),
			huh.NewText().Title("Description").Value(&bookmark.Description),
		)).Run()

		bookmark.Tags = strings.Split(tags, ",")

		err = store.UpdateBookmark(db, original_bookmark, bookmark)
		if err != nil {
			fmt.Println(err.Error())
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// editCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// editCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
