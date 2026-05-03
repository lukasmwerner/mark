/*
Copyright © 2025 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"

	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// keysDeleteCmd represents the delete command
var keysDeleteCmd = &cobra.Command{
	Use:   "delete <key>",
	Short: "Delete a key",
	Long:  `Pass in the key to delete from the allowed api keys for the local http server`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open(store.Options{
			Flags: []store.Flag{store.Embedding},
		})
		if err != nil {
			fmt.Println(err)
			return
		}
		defer db.Close()

		err = store.DeleteKey(db, args[0])
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println("Key deleted successfully")
	},
}

func init() {
	keysCmd.AddCommand(keysDeleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
