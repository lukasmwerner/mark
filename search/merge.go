package search

import (
	"sort"

	"github.com/lukasmwerner/mark/store"
)

type Source string

const (
	FTS        = Source("FTS")
	PartialFTS = Source("Partial FTS")
	Embedding  = Source("Embedding")
)

func MergeResults(sources []Source, results ...[]store.Bookmark) []Result {
	type rankedBookmark struct {
		bookmark store.Bookmark
		score    float64
		bestRank int
		firstSet int
		source   Source
	}

	merged := map[string]*rankedBookmark{}
	for setIndex, resultSet := range results {
		sourceWeight := 1.0
		for range setIndex {
			sourceWeight *= 0.25
		}

		for rank, bm := range resultSet {
			key := bm.Url
			if key == "" {
				key = bm.Title
			}

			// Weighted reciprocal rank fusion: bookmarks that rank highly in one or more
			// result sets bubble toward the front of the merged list, with each source's
			// influence halved by source order: 1, 0.5, 0.25, ...
			score := sourceWeight / float64(rank+1)
			if existing, ok := merged[key]; ok {
				existing.score += score
				if rank < existing.bestRank {
					existing.bestRank = rank
				}
				continue
			}

			merged[key] = &rankedBookmark{
				bookmark: bm,
				score:    score,
				bestRank: rank,
				firstSet: setIndex,
				source:   sources[setIndex],
			}
		}
	}

	outputResults := make([]rankedBookmark, 0, len(merged))
	for _, bm := range merged {
		outputResults = append(outputResults, *bm)
	}

	sort.SliceStable(outputResults, func(i, j int) bool {
		if outputResults[i].score != outputResults[j].score {
			return outputResults[i].score > outputResults[j].score
		}
		if outputResults[i].bestRank != outputResults[j].bestRank {
			return outputResults[i].bestRank < outputResults[j].bestRank
		}
		if outputResults[i].firstSet != outputResults[j].firstSet {
			return outputResults[i].firstSet < outputResults[j].firstSet
		}
		return outputResults[i].bookmark.Url < outputResults[j].bookmark.Url
	})

	bookmarks := make([]Result, len(outputResults))
	for i, bm := range outputResults {
		bookmarks[i] = Result{
			Bookmark: bm.bookmark,
			Source:   bm.source,
			Rank:     bm.score,
		}
	}

	return bookmarks
}
