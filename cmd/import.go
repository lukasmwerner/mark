/*
Copyright © 2025 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import <file>",
	Short: "Lets you import bookmarks from an existing csv",
	Long: `Import bookmarks from an existing csv file.

csv format:
title,description,tags,url
"Title","Description","tag1,tag2","https://example.com",
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		f, err := os.Open(args[0])
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		db, err := store.Open(store.Options{
			Flags: []store.Flag{store.Embedding},
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		r := csv.NewReader(f)

		for {
			records, err := r.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
			}
			bm := store.Bookmark{
				Title:       records[0],
				Description: records[1],
				Tags:        strings.Split(records[2], ","),
				Url:         records[3],
			}
			id, err := store.InsertBookmark(db, bm)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			fmt.Println(id, "-", bm.Url)
		}
	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// importCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
