/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/lukasmwerner/mark/store"
	"github.com/spf13/cobra"
)

// createEmbeddingsCmd represents the createEmbeddings command
var createEmbeddingsCmd = &cobra.Command{
	Use:   "createEmbeddings",
	Short: "Generates vector embeddings that are not currently in the embeddings table",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open(store.Options{
			Flags: []store.Flag{store.Embedding},
		})
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		res, err := db.Exec(`INSERT INTO
		bookmark_embeddings (document_id, embedding)
		SELECT b.id, embed('embeddinggemma', concat_ws(' ', 'title: ', b.title, ' | text: ', b.description, b.tags, justPath(b.url)))
		FROM Bookmarks b
		WHERE NOT EXISTS (
			SELECT 1 FROM bookmark_embeddings be WHERE be.document_id = b.id
		);`)
		if err != nil {
			log.Fatalln(err.Error())
		}
		rows, _ := res.RowsAffected()
		fmt.Sprintf("Successfully did %s rows\n", rows)
	},
}

func init() {
	rootCmd.AddCommand(createEmbeddingsCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createEmbeddingsCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createEmbeddingsCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
