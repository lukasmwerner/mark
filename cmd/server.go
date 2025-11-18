/*
Copyright © 2024 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"

	mark_http "github.com/lukasmwerner/mark/http"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Local HTTP server for managing bookmarks",
	Long:  `Designed for hosting for applications where there is no strong storage api that can easily be synchronized with Dropbox, Google Drive, Syncthing or other cloud storage sync services.`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open()
		if err != nil {
			fmt.Println("error occured in opening db: ", err.Error())
			return
		}

		mark_http.RegisterRoutes(db, http.DefaultServeMux)

		log.Fatal(http.ListenAndServe(":1990", nil))
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
