package store

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/mattn/go-sqlite3"
)

type requirement struct {
	name       string
	definition string
}

var Tables = []requirement{
	{
		name: "Bookmarks",
		definition: `CREATE TABLE IF NOT EXISTS Bookmarks (
    id INTEGER PRIMARY KEY NOT NULL,
    url TEXT,
    title TEXT,
    description TEXT,
    tags TEXT
);`,
	},
	{
		name: "Bookmarks_fts",
		definition: `CREATE VIRTUAL TABLE IF NOT EXISTS Bookmarks_fts USING fts5(
    url,
    title,
    description,
    tags
);`,
	},
	{
		name: "Bookmark_Sync_1",
		definition: `CREATE TRIGGER IF NOT EXISTS Bookmarks_insert AFTER INSERT ON Bookmarks
BEGIN
    INSERT INTO Bookmarks_fts (rowid, url, title, description, tags)
    VALUES (new.id, new.url, new.title, new.description, new.tags);
END;`,
	},
	{
		name: "Bookmark_Sync_2",
		definition: `CREATE TRIGGER IF NOT EXISTS Bookmarks_delete AFTER DELETE ON Bookmarks
BEGIN
    DELETE FROM Bookmarks_fts WHERE rowid = old.id;
END;`,
	},
	{
		name: "Bookmark_Sync_3",
		definition: `CREATE TRIGGER IF NOT EXISTS Bookmarks_update AFTER UPDATE ON Bookmarks
BEGIN
    DELETE FROM Bookmarks_fts WHERE rowid = old.id;
    INSERT INTO Bookmarks_fts (rowid, url, title, description, tags)
    VALUES (new.id, new.url, new.title, new.description, new.tags);
END;`,
	},
	{
		name: "Server_Keys",
		definition: `CREATE TABLE IF NOT EXISTS Server_Keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    key TEXT NOT NULL UNIQUE
);`,
	},
	{
		name:       "Preformance_Tune_WAL",
		definition: `PRAGMA journal_mode=WAL;`,
	},
}

func Open() (*DB, error) {
	markStoreLocation := os.Getenv("MARK_STORE_LOCATION")
	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, errors.Join(errors.New("unable to get homedir"), err)
	}
	if markStoreLocation == "" {
		markStoreLocation = path.Join(homedir, ".config", "mark")
	}

	if err := EnsureDirExists(markStoreLocation); err != nil {
		return nil, errors.Join(errors.New("unable to make mark store location in: "+markStoreLocation), err)
	}
	changesPath := path.Join(markStoreLocation, "changes")
	if err := EnsureDirExists(changesPath); err != nil {
		return nil, errors.Join(errors.New("unable to make mark store changes location in: "+markStoreLocation), err)
	}

	var ext string
	switch runtime.GOOS {
	case "windows":
		ext = ".dll"
	case "darwin":
		ext = ".dylib"
	default:
		ext = ".so"
	}

	sql.Register("cr-sqlite", &sqlite3.SQLiteDriver{
		Extensions: []string{path.Join(homedir, ".lib", "crsqlite"+ext)},
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
	if err != nil {
		log.Println(err.Error())
	}

	_, err = db.Exec("select crsql_as_crr('Bookmarks');")
	if err != nil {
		return nil, errors.Join(errors.New("unable to setup crdts"), err)
	}

	err = syncronizeFromHostsToDB(db, hostname, changesPath)
	if err != nil {
		return nil, errors.Join(errors.New("unable to sync fs -> db"), err)
	}

	return db, nil
}

type DB struct {
	*sql.DB

	StoreLoc        string
	ChangesStoreLoc string
	Hostname        string
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

func EnsureTables(db *DB, tables ...requirement) error {
	for _, table := range tables {
		_, err := db.Exec(table.definition)
		if err != nil {
			log.Print(table.name, err.Error())
			return err
		}
	}

	return nil
}

func InsertBookmark(db *DB, bookmark Bookmark) (BookmarkId, error) {
	tags := strings.Join(bookmark.Tags, ", ")
	result, err := db.Exec("INSERT INTO Bookmarks (url, title, description, tags) VALUES (?, ?, ?, ?)",
		bookmark.Url, bookmark.Title, bookmark.Description, tags)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	return BookmarkId(id), err
}

func GetBookmark(db *DB, query_url string) (Bookmark, error) {
	var b Bookmark
	var tags string

	u, err := url.Parse(query_url)
	if err != nil {
		return b, err
	}

	rows, err := db.Query("SELECT url, title, description, tags FROM Bookmarks WHERE url like ? AND url like ?", fmt.Sprintf("%%%v%%", u.Host), fmt.Sprintf("%%%v%%", u.Path))
	if err != nil {
		return b, err
	}

	i := 0
	for rows.Next() {
		if i > 0 {
			return b, errors.New("multiple bookmarks found")
		}
		err := rows.Scan(&b.Url, &b.Title, &b.Description, &tags)
		if err != nil {
			return b, err
		}
		b.Tags = strings.Split(tags, ", ")
		i++
	}

	if i == 0 {
		return b, sql.ErrNoRows
	}

	db_url, err := url.Parse(b.Url)
	if err != nil {
		return b, errors.New("unable to parse bookmark url")
	}

	if db_url.Path != u.Path {
		return b, sql.ErrNoRows
	}

	return b, nil
}

func SearchBookmarks(db *DB, query string) ([]Bookmark, error) {
	bookmarks := []Bookmark{}
	rows, err := db.Query(`SELECT url, title, description, tags FROM Bookmarks_fts WHERE Bookmarks_fts MATCH ?;`, query)
	if err != nil {
		return bookmarks, err
	}
	defer rows.Close()

	for rows.Next() {
		var b Bookmark
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

func UpdateBookmark(db *DB, original Bookmark, updated Bookmark) error {
	_, err := db.Exec(`UPDATE Bookmarks SET
		url = ?,
		title = ?,
		description = ?,
		tags = ?
	WHERE
		url = ?;`,
		updated.Url,
		updated.Title,
		updated.Description,
		strings.Join(updated.Tags, ", "),
		original.Url,
	)

	return err
}

func AddKey(db *DB, key string) error {
	_, err := db.Exec("INSERT INTO Server_Keys (key) VALUES (?)", key)
	if err != nil {
		return err
	}
	return nil
}

func HasKey(db *DB, key string) (bool, error) {
	var value string
	err := db.QueryRow("SELECT key FROM Server_Keys WHERE key = ?", key).Scan(&value)
	if err != nil || value == "" {
		return false, err
	}
	return true, nil
}

func GetKeys(db *DB) ([]string, error) {
	keys := []string{}
	rows, err := db.Query("SELECT key FROM Server_Keys")
	if err != nil {
		return keys, err
	}
	defer rows.Close()

	for rows.Next() {
		var key string
		err := rows.Scan(&key)
		if err != nil {
			return keys, err
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func DeleteKey(db *DB, key string) error {
	_, err := db.Exec("DELETE FROM Server_Keys WHERE key = ?", key)
	if err != nil {
		return err
	}
	return nil
}
