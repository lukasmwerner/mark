/*
Copyright © 2025 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"database/sql"
	"fmt"

	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "dumps the database to a sqlite3 file",
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

		dumpDB, err := sql.Open("sqlite3", args[0])
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		defer dumpDB.Close()

		dumpDB.Exec(`CREATE TABLE bookmarks (
    id INTEGER PRIMARY KEY NOT NULL,
    url TEXT,
    title TEXT,
    description TEXT,
    tags TEXT
);`)

		rows, err := db.QueryContext(cmd.Context(), "SELECT url, title, description, tags FROM Bookmarks;")
		if err != nil {
			fmt.Println(err.Error())
			return
		}

		defer rows.Close()

		for rows.Next() {
			var b store.Bookmark
			var tags string
			err := rows.Scan(&b.Url, &b.Title, &b.Description, &tags)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
			_, err = dumpDB.Exec(`INSERT INTO bookmarks (url, title, description, tags) VALUES (?, ?, ?, ?)`,
				b.Url, b.Title, b.Description, tags)
			if err != nil {
				fmt.Println(err.Error())
				return
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(dumpCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// dumpCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// dumpCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
