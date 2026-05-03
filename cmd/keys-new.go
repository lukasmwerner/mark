/*
Copyright © 2024 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// newKeyCmd represents the server command
var newKeyCmd = &cobra.Command{
	Use:   "new",
	Short: "Generate a new key",
	Long:  `Makes a new key for use in the mark http server (web app, web extension and raycast-extension)`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open(store.Options{
			Flags: []store.Flag{store.Embedding},
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()

		key := generateKey()
		err = store.AddKey(db, key)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("key: ", key)
	},
}

func init() {
	keysCmd.AddCommand(newKeyCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func generateKey() string {
	return fmt.Sprintf("%x", uuid.New().String())
}
