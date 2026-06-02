/*
Copyright © 2025 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/lukasmwerner/mark/search"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "deletes bookmarks based on given search query",
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
		bookmarks := search.ToBookmarks(results)
		for _, bm := range bookmarks {
			var affirmative bool
			err := huh.NewConfirm().Title("Delete?").
				Description(prettyOutput(bm)).
				Affirmative("Delete!").
				Negative("No, Keep").
				Value(&affirmative).Run()
			if err != nil {
				return
			}
			if !affirmative {
				fmt.Println("Skipped.")
				continue
			}
			err = store.DeleteBookmark(db, bm)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("Deleted.")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
