package search

import "github.com/lukasmwerner/mark/store"

type Result struct {
	store.Bookmark
	Source Source
	Rank   float64
}
