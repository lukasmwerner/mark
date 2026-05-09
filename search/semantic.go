package search

import (
	"strings"

	"github.com/lukasmwerner/mark/store"
)

type semanticResult struct {
	bm    store.Bookmark
	score float64
}

func (s semanticResult) Get() store.Bookmark {
	return s.bm
}
func (s semanticResult) Score() float64 {
	return s.score
}

func Semantic(db *store.DB, query string) ([]Bookmark, error) {
	var results []Bookmark
	rows, err := db.Query(`SELECT b.url, b.title, b.description, b.tags, distance
				FROM bookmark_embeddings e
				JOIN Bookmarks b ON e.document_id = b.id
				WHERE e.embedding MATCH embed('embeddinggemma', concat_ws(' ', 'task: search result | query: ', ?)) and k = 100 and distance <= 1.21
				ORDER BY distance;`, query)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var b store.Bookmark
		var tags string
		var distance float64
		err := rows.Scan(&b.Url, &b.Title, &b.Description, &tags, &distance)
		if err != nil {
			return results, err
		}
		b.Tags = strings.Split(tags, ", ")
		results = append(results, semanticResult{
			bm:    b,
			score: distance,
		})
	}

	return results, nil
}
