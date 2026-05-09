package search

import (
	"strings"

	"github.com/lukasmwerner/mark/store"
)

type fullTextResult struct {
	bm    store.Bookmark
	score float64
}

func (f fullTextResult) Get() store.Bookmark {
	return f.bm
}
func (f fullTextResult) Score() float64 {
	return f.score
}

func FullText5(db *store.DB, query string) ([]Bookmark, error) {
	var results []fullTextResult
	var maximum, minimum float64
	first := true

	rows, err := db.Query(`SELECT url, title, description, tags, bm25(Bookmarks_fts)
		FROM Bookmarks_fts
		WHERE Bookmarks_fts MATCH ?
		ORDER BY bm25(Bookmarks_fts);`, query)
	if err != nil {
		return []Bookmark{}, err
	}
	defer rows.Close()

	for rows.Next() {
		var b store.Bookmark
		var tags string
		var score float64
		err := rows.Scan(&b.Url, &b.Title, &b.Description, &tags, &score)
		if err != nil {
			return []Bookmark{}, err
		}
		b.Tags = strings.Split(tags, ", ")

		if first {
			maximum = score
			minimum = score
			first = false
		} else {
			if score > maximum {
				maximum = score
			} else if score < minimum {
				minimum = score
			}
		}

		results = append(results, fullTextResult{
			bm:    b,
			score: score,
		})
	}

	out := make([]Bookmark, len(results))
	for i, b := range results {
		results[i].score = 1 - (b.score-minimum)/(maximum-minimum)
		out[i] = results[i]
	}
	return out, nil
}

func PartialText(db *store.DB, query string) ([]Bookmark, error) {
	var results []Bookmark

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
		ORDER BY bm25(Bookmarks_fts);`,
		fuzzy_query)

	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var b store.Bookmark
		var tags string
		var id int
		err := rows.Scan(&b.Url, &b.Title, &b.Description, &id, &tags)
		if err != nil {
			return results, err
		}
		b.Tags = strings.Split(tags, ", ")
		results = append(results, fullTextResult{bm: b})
	}

	return results, nil
}
