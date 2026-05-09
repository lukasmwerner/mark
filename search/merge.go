package search

import (
	"sort"

	"github.com/lukasmwerner/mark/store"
)

type Source string

const (
	FTS        = Source("FTS")
	PartialFTS = Source("pFTS")
	Embedding  = Source("Embed")
)

var sourceWeights = map[Source]float64{
	FTS:        1.0,
	Embedding:  0.7,
	PartialFTS: 0.4,
}

func MergeResults(sources []Source, results ...[]Bookmark) []Result {
	type rankedBookmark struct {
		bookmark store.Bookmark
		score    float64
		boost    float64
		bestRank int
		firstSet int
		source   Source
	}

	merged := map[string]*rankedBookmark{}
	for setIndex, resultSet := range results {
		sourceKind := sources[setIndex]
		sourceWeight := sourceWeights[sourceKind]

		for rank, b := range resultSet {
			bm := b.Get()
			key := bm.Url

			// Weighted reciprocal rank fusion: bookmarks that rank highly in one or more
			// result sets bubble toward the front of the merged list, with each source's
			// influence determined by their weight in the sourceWeights map
			score := sourceWeight / float64(rank+1)

			// Score boosting depending on different source heuristics
			boost := 1.0
			switch sourceKind {
			case Embedding:
				boost = embeddingMultiplier(b)
			case FTS:
				score = sourceWeight * b.Score()
			default:
			}
			score *= boost

			if existing, ok := merged[key]; ok {
				if boost != 1.0 {
					existing.boost += boost
				}
				existing.score += score // Boost things that are duplicate
				if rank < existing.bestRank {
					existing.bestRank = rank
				}
				continue
			}

			merged[key] = &rankedBookmark{
				bookmark: bm,
				score:    score,
				boost:    boost,
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
