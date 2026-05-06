package search

import "github.com/lukasmwerner/mark/store"

func Search(db *store.DB, query string) ([]Result, error) {
	results := []Result{}
	// TODO: make these searches concurrent?
	fts5, err := FullText5(db, query)
	if err != nil {
		return results, err
	}
	partials, err := PartialText(db, query)
	if err != nil {
		return results, err
	}
	semantic, err := Semantic(db, query)
	if err != nil {
		return results, err
	}

	results = MergeResults([]Source{FTS, Embedding, PartialFTS}, fts5, semantic, partials)

	return results, nil
}
