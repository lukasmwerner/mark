package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
)

type crsql_changes struct {
	Table       string
	Pk          []byte
	Cid         string
	Value       []byte
	Col_version int
	Db_version  int
	Site_id     []byte
	Cl          int
	Seq         int
}

func syncronizeLocalChangesToDisk(db *DB, hostFile string) error {
	rows, err := db.Query("SELECT * FROM crsql_changes where site_id = crsql_site_id();")
	if err != nil {
		return err
	}
	changes := []crsql_changes{}
	for rows.Next() {
		var table string
		var pk []byte
		var cid string
		var value []byte
		var col_version int
		var db_version int
		var site_id []byte
		var cl int
		var seq int
		err := rows.Scan(&table, &pk, &cid, &value, &col_version, &db_version, &site_id, &cl, &seq)
		if err != nil {
			return err
		}
		changes = append(changes, crsql_changes{
			Table:       table,
			Pk:          pk,
			Cid:         cid,
			Value:       value,
			Col_version: col_version,
			Db_version:  db_version,
			Site_id:     site_id,
			Cl:          cl,
			Seq:         seq,
		})
	}
	f, err := os.Create(hostFile + ".changes")
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.Marshal(&changes)
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	if err != nil {
		return err
	}

	return nil
}



func syncronizeFromHostsToDB(db *DB, hostname, changesPath string) error {

	// Synchronize any new changes
	hosts, err := os.ReadDir(changesPath)
	if err != nil {
		return err
	}
	for _, host := range hosts {
		if host.Name() == hostname || path.Ext(host.Name()) != ".changes" {
			continue
		}
		if host.IsDir() {
			continue
		}
		err := syncronizeFromDiskToDB(db, path.Join(changesPath, host.Name()))
		if err != nil {
			return errors.Join(fmt.Errorf("sync disk -> db: %s", host.Name()), err)
		}
	}

	return nil
}

func syncronizeFromDiskToDB(db *DB, hostFile string) error {

	f, err := os.Open(hostFile)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	var changes []crsql_changes
	err = json.Unmarshal(b, &changes)
	if err != nil {
		return err
	}
	for _, change := range changes {
		_, err := db.Exec("INSERT INTO crsql_changes VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			change.Table,
			change.Pk,
			change.Cid,
			change.Value,
			change.Col_version,
			change.Db_version,
			change.Site_id,
			change.Cl,
			change.Seq,
		)
		if err != nil {
			return err
		}

	}

	return nil
}
