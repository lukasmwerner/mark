package search

import "github.com/lukasmwerner/mark/store"

type Result struct {
	store.Bookmark
	Source Source
	Rank   float64
}

type Bookmark interface {
	Get() store.Bookmark
	Score() float64
}

func ToBookmarks(results []Bookmark) []store.Bookmark {
	res := make([]store.Bookmark, len(results))
	for i, b := range results {
		res[i] = b.Get()
	}
	return res
}
