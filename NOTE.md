# Spellfix1

missing module from `mattn/sqlite`, need to use https://github.com/mattn/go-sqlite3/issues/1160 

1. ensure spellfix is available (likely compile tags)

2. have bookmark table
    - `CREATE VIRTUAL TABLE IF NOT EXISTS Bookmarks_vocabulary USING spellfix1`
3. populate vocab table with all bookmarks (title, tags, description)
    - will need a 'migration' capability
4. add better ui info showing that we are doing a multi-stage setup process
    - [ ] possibly stream in the results if it doesn't happen in time?


Example:
```
-- User searches: "javascrpt tutorial"
-- 1. FTS5 search (no results)
SELECT * FROM Bookmarks_fts WHERE Bookmarks_fts MATCH 'javascrpt tutorial';
-- 2. Spellfix1 correction
SELECT word FROM spellfix WHERE word MATCH 'javascrpt';  -- returns "javascript"
-- 3. Retry FTS5 with corrected query
SELECT * FROM Bookmarks_fts WHERE Bookmarks_fts MATCH 'javascript tutorial';
```


```go
func CorrectSpelling(db *DB, query string) string {
    words := strings.Fields(query)
    correctedWords := make([]string, len(words))

    for i, word := range words {
        // Skip operators
        if word == "AND" || word == "OR" || word == "NOT" {
            correctedWords[i] = word
            continue
        }

        // Try to find spelling correction
        var suggestion string
        err := db.QueryRow(`
            SELECT word FROM Bookmarks_vocabulary
            WHERE word MATCH ?
            AND top = 1
            LIMIT 1;
        `, word).Scan(&suggestion)

        if err == nil && suggestion != "" {
            correctedWords[i] = suggestion
        } else {
            correctedWords[i] = word
        }
    }
    return strings.Join(correctedWords, " ")
}
```


pseudocode:

```go
func AutoCorrectingSearch(db *DB, query string) ([]Bookmark, error) {
    // Tier 1: Try exact FTS match
    results, err := SearchBookmarks(db, query)
    if err == nil && len(results) > 0 {
        return results, nil
    }

    // Tier 2: Try spell correction
    correctedQuery := CorrectSpelling(db, query)

    // Only retry if correction actually changed something
    if correctedQuery != query {
        results, err = SearchBookmarks(db, correctedQuery)
        if err == nil && len(results) > 0 {
            return results, nil
        }
    }

    // Tier 3: do vector search matching

    // Return whatever we have (may be empty)
    return results, err
}
```
