/*
Copyright © 2025 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"

	_ "embed"

	"github.com/a-h/templ"
	"github.com/lukasmwerner/mark/search"
	"github.com/lukasmwerner/mark/store"
	"github.com/lukasmwerner/mark/web"
	"github.com/lukasmwerner/mark/web/static"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: `[EXPERIMENTAL] google search like interface`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open(store.Options{
			Flags: []store.Flag{store.Embedding},
		})
		if err != nil {
			log.Println(err.Error())
			return
		}
		go db.FSWatcher()

		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))
		http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query().Get("q")
			if q == "" {
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
			results, err := search.Search(db, q)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "oops we had something go wrong.")
				fmt.Fprintln(w, err.Error())
				return
			}
			templ.Handler(web.ResultsPage("lukaswerner.com", q, "(FTS + Embeddings) RRF", results)).ServeHTTP(w, r)
		})
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			n := store.CountBookmarks(db)
			templ.Handler(web.LandingPage("lukaswerner.com", n)).ServeHTTP(w, r)
		})

		err = http.ListenAndServe(":1995", nil)
		if err != nil {
			log.Println(err.Error())
			return
		}

	},
}

func init() {
	rootCmd.AddCommand(searchCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// searchTuiCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// searchTuiCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
