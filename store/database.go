package store

import (
	"database/sql"
	"errors"
	"os"
	"path"

	"github.com/mattn/go-sqlite3"
)

type table struct {
	name       string
	definition string
}

var Tables = []table{
	{
		name: "Bookmarks",
		definition: `CREATE TABLE IF NOT EXISTS Bookmarks (
	bookmark_id INTEGER PRIMARY KEY AUTOINCREMENT,
	url TEXT NOT NULL UNIQUE,
	title TEXT NOT NULL,
	description TEXT,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`,
	},
	{
		name: "Tags",
		definition: `CREATE TABLE IF NOT EXISTS Tags (
		tag_id INTEGER PRIMARY KEY AUTOINCREMENT,
		tag_name TEXT NOT NULL UNIQUE
	);`,
	},
	{
		name: "BookmarkTags",
		definition: `CREATE TABLE BookmarkTags (
		bookmark_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		PRIMARY KEY (bookmark_id, tag_id),
		FOREIGN KEY (bookmark_id) REFERENCES Bookmarks (bookmark_id),
		FOREIGN KEY (tag_id) REFERENCES Tags (tag_id)
	);`,
	},
}

func Open() (*DB, error) {
	markStoreLocation := os.Getenv("MARK_STORE_LOCATION")
	if markStoreLocation == "" {
		homedir, err := os.UserHomeDir()
		if err != nil {
			return nil, errors.Join(errors.New("unable to get homedir"), err)
		}
		markStoreLocation = path.Join(homedir, ".config", "mark")
	}

	if err := EnsureDirExists(markStoreLocation); err != nil {
		return nil, errors.Join(errors.New("unable to make mark store location in: "+markStoreLocation), err)
	}
	changesPath := path.Join(markStoreLocation, "changes")
	if err := EnsureDirExists(changesPath); err != nil {
		return nil, errors.Join(errors.New("unable to make mark store changes location in: "+markStoreLocation), err)
	}

	sql.Register("cr-sqlite", &sqlite3.SQLiteDriver{
		Extensions: []string{"crsqlite"},
	})

	sqlDB, err := sql.Open("cr-sqlite", path.Join(markStoreLocation, "data.db"))
	if err != nil {
		return nil, errors.Join(errors.New("unable to open database"), err)
	}

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	db := &DB{
		DB:              sqlDB,
		StoreLoc:        markStoreLocation,
		ChangesStoreLoc: changesPath,
		Hostname:        hostname,
	}

	err = EnsureTables(db, Tables...)

	if err == nil { // Tables were setup so we need to finish
		for _, table := range Tables {
			_, err = db.Exec("select crsql_as_crr('"+table.name+"');")
			if err != nil {
				return nil, errors.Join(errors.New("unable to open database"), err)
			}
		}
	}

	err = syncronizeFromHostsToDB(db, hostname, changesPath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

type DB struct {
	*sql.DB

	StoreLoc string
	ChangesStoreLoc string
	Hostname string
}

func (db *DB) Close() error {
	err := syncronizeLocalChangesToDisk(db, path.Join(db.ChangesStoreLoc, db.Hostname))
	if err != nil {
		return err
	}

	_, err = db.Exec(`select crsql_finalize();`) // Clean up after cr-sqlite
	if err != nil {
		return err
	}

	return db.DB.Close()
}

func EnsureTables(db *DB, tables ...table) error {
	for _, table := range tables {
		_, err := db.Exec(table.definition)
		if err != nil {
			return err
		}
	}

	return nil
}


func InsertBookmark(db *DB, bookmark Bookmark) (BookmarkId, error) {
    result, err := db.Exec("INSERT INTO Bookmarks (url, title, description) VALUES (?, ?, ?)",
        bookmark.Url, bookmark.Title, bookmark.Description)
    if err != nil {
        return 0, err
    }
	id, err := result.LastInsertId()
	return BookmarkId(id), err
}


func InsertTagsAndAssociate(db *DB, bookmarkID BookmarkId, tags []string) error {
    // Insert tags and get their IDs
    tagIDs := make(map[string]int64)
    for _, tag := range tags {
        var tagID int64
        err := db.QueryRow("SELECT tag_id FROM Tags WHERE tag_name = ?", tag).Scan(&tagID)
        if err == sql.ErrNoRows {
            // Insert the tag if it doesn't exist
            result, err := db.Exec("INSERT INTO Tags (tag_name) VALUES (?)", tag)
            if err != nil {
                return err
            }
            tagID, err = result.LastInsertId()
            if err != nil {
                return err
            }
        } else if err != nil {
            return err
        }
        tagIDs[tag] = tagID
    }

    // Associate the tags with the bookmark
    for _, tag := range tags {
        _, err := db.Exec("INSERT INTO BookmarkTags (bookmark_id, tag_id) VALUES (?, ?)", bookmarkID, tagIDs[tag])
        if err != nil {
            return err
        }
    }

    return nil
}



