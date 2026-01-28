/*
Copyright © 2025 Lukas Werner <me@lukaswerner.com>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	_ "embed"

	"github.com/a-h/templ"
	"github.com/lukasmwerner/mark/store"
	"github.com/lukasmwerner/mark/web"
	"github.com/lukasmwerner/mark/web/static"
	"github.com/spf13/cobra"
)

type RankingMethod string

const (
	recency RankingMethod = "rowid DESC, rank DESC" // rank based on rowid will be an issue when going across users
	ftsRank RankingMethod = "rank DESC, rowid DESC"
)

func WebSearchBookmarks(db *store.DB, query string, ranker RankingMethod) ([]store.Bookmark, error) {
	bookmarks := []store.Bookmark{}

	partials := ""
	for _, field := range strings.Fields(query) {
		partial := ""
		switch field {
		case "NOT":
			partial = field
		case "OR":
			partial = field
		case "AND":
			partial = field
		default:
			partial = field + "*"

		}
		partials += " " + partial
	}

	rows, err := db.Query(
		`
		SELECT
			url,
			title,
			description,
			rowid,
			tags
		FROM (
		SELECT
			url,
			title,
			description,
			tags,
			rowid,
			rank,
			1 AS priority
		FROM Bookmarks_fts
		WHERE Bookmarks_fts MATCH ?

		UNION

		SELECT
			url,
			title,
			description,
			tags,
			rowid,
			rank,
			2 AS priority
		FROM Bookmarks_fts
		WHERE Bookmarks_fts MATCH ?)

		GROUP BY url
		ORDER BY MIN(priority) ASC, `+string(ranker)+`;`,
		query, partials)

	// rows, err := db.Query(
	// 	`SELECT url, title, description, tags FROM Bookmarks_fts WHERE Bookmarks_fts MATCH ? ORDER BY rowid DESC, rank DESC;`,
	// 	query)
	if err != nil {
		return bookmarks, err
	}
	defer rows.Close()

	for rows.Next() {
		var b store.Bookmark
		var tags string
		var id int
		err := rows.Scan(&b.Url, &b.Title, &b.Description, &id, &tags)
		if err != nil {
			return bookmarks, err
		}
		b.Tags = strings.Split(tags, ", ")
		bookmarks = append(bookmarks, b)
	}

	return bookmarks, nil
}

var searchCmd = &cobra.Command{
	Use:   "search",
	Short: `[EXPERIMENTAL] google search like interface`,
	Run: func(cmd *cobra.Command, args []string) {
		db, err := store.Open()
		if err != nil {
			log.Println(err.Error())
			return
		}

		http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static.FS))))
		http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query().Get("q")
			if q == "" {
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
			unsafe_order := r.URL.Query().Get("order")
			order := recency
			orderBy := "Recency"
			if unsafe_order == string("fts") {
				order = ftsRank
				orderBy = "FTS"
			}
			results, err := WebSearchBookmarks(db, q, order)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintln(w, "oops we had something go wrong.")
				fmt.Fprintln(w, err.Error())
				return
			}
			templ.Handler(web.ResultsPage("lukaswerner.com", q, orderBy, results)).ServeHTTP(w, r)
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
