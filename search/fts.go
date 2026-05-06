package search

import (
	"strings"

	"github.com/lukasmwerner/mark/store"
)

func FullText5(db *store.DB, query string) ([]store.Bookmark, error) {
	bookmarks := []store.Bookmark{}

	rows, err := db.Query(`SELECT url, title, description, tags
		FROM Bookmarks_fts
		WHERE Bookmarks_fts MATCH ?
		ORDER BY bm25(Bookmarks_fts) DESC;`, query)
	if err != nil {
		return bookmarks, err
	}
	defer rows.Close()

	for rows.Next() {
		var b store.Bookmark
		var tags string
		err := rows.Scan(&b.Url, &b.Title, &b.Description, &tags)
		if err != nil {
			return bookmarks, err
		}
		b.Tags = strings.Split(tags, ", ")
		bookmarks = append(bookmarks, b)
	}

	return bookmarks, nil
}

func PartialText(db *store.DB, query string) ([]store.Bookmark, error) {
	bookmarks := []store.Bookmark{}

	fuzzy_query := ""
	for _, field := range strings.Fields(query) {
		fragment := ""
		switch field {
		case "NOT":
			fragment = field
		case "OR":
			fragment = field
		case "AND":
			fragment = field
		default:
			fragment = field + "*"

		}
		fuzzy_query += " " + fragment
	}

	rows, err := db.Query(`SELECT
			url,
			title,
			description,
			rowid,
			tags
		FROM Bookmarks_fts
		WHERE Bookmarks_fts MATCH ?
		ORDER BY bm25(Bookmarks_fts) DESC;`,
		fuzzy_query)

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
