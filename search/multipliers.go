package search

func embeddingMultiplier(b Bookmark) float64 {
	distance := b.Score()

	const strongCutoff = 1.0
	const semanticCutoff = 1.21
	const maxBoost = 2.0
	const weakMaxMultiplier = 0.25

	if distance < strongCutoff {
		closeness := strongCutoff - distance
		return 1.0 + maxBoost*closeness
	}

	if distance >= semanticCutoff {
		return 0.0
	}

	stretch := distance - strongCutoff
	stretchRange := semanticCutoff - strongCutoff
	return weakMaxMultiplier * (1.0 - stretch/stretchRange)
}
