package search

import (
	"strings"

	"github.com/lukasmwerner/mark/store"
)

func Semantic(db *store.DB, query string) ([]store.Bookmark, error) {
	bookmarks := []store.Bookmark{}
	rows, err := db.Query(`SELECT b.url, b.title, b.description, b.tags
				FROM bookmark_embeddings e
				JOIN Bookmarks b ON e.document_id = b.id
				WHERE e.embedding MATCH embed('embeddinggemma', concat_ws(' ', 'task: search result | query: ', ?)) and k = 100 and distance <= 1.21
				ORDER BY distance;`, query)
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
