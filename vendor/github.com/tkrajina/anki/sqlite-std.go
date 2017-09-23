// +build !js

// Copyright: Jonathan Hall
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html

package anki

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sqlx.DB
	tmpFile string
}

func (db *DB) Collection() (*Collection, error) {
	var deletedDecks []ID
	if rows, err := db.Query("SELECT oid FROM graves WHERE type=2"); err != nil {
		return nil, err
	} else {
		for rows.Next() {
			id := new(ID)
			if err := rows.Scan(id); err != nil {
				return nil, err
			}
			deletedDecks = append(deletedDecks, *id)
		}
	}
	collection := &Collection{}
	if err := db.Get(collection, "SELECT * FROM col"); err != nil {
		return nil, err
	}
	for _, deck := range collection.Decks {
		for _, deleted := range deletedDecks {
			if deck.ID == deleted {
				delete(collection.Decks, deck.ID)
				continue
			}
		}
		conf, ok := collection.DeckConfigs[deck.ConfigID]
		if !ok {
			//return nil, fmt.Errorf("Deck %d references non-existent config %d", deck.ID, deck.ConfigID)
		}
		deck.Config = conf
	}
	return collection, nil
}

func (db *DB) Cards() (*Cards, error) {
	rows, err := db.Queryx(`
		SELECT c.id, c.nid, c.did, c.ord, c.mod, c.usn, c.type, c.queue, c.reps, c.lapses, c.left, c.odid,
			CAST(c.factor AS real)/1000 AS factor,
			CASE c.queue
				WHEN 0 THEN NULL
				WHEN 1 THEN c.due
				WHEN 2 THEN c.due*24*60*60+(SELECT crt FROM col)
			END AS due,
			CASE
				WHEN c.ivl == 0 THEN NULL
				WHEN c.ivl < 0 THEN -ivl
				ELSE c.ivl*24*60*60
			END AS ivl,
			CASE c.queue
				WHEN 0 THEN NULL
				WHEN 1 THEN c.odue
				WHEN 2 THEN c.odue*24*60*60+(SELECT crt FROM col)
			END AS odue
		FROM cards c
		LEFT JOIN graves g ON g.oid=c.id AND g.type=0
		WHERE g.oid IS NULL
		ORDER BY id DESC
	`)
	return &Cards{rows}, err
}

func (db *DB) Notes() (*Notes, error) {
	rows, err := db.Queryx(`
		SELECT n.id, n.guid, n.mid, n.mod, n.usn, n.tags, n.flds, n.sfld,
			CAST(n.csum AS text) AS csum -- Work-around for SQL.js trying to treat this as a float
		FROM notes n
		LEFT JOIN graves g ON g.oid=n.id AND g.type=1
		ORDER BY id DESC
	`)
	return &Notes{rows}, err
}

func (db *DB) Reviews() (*Reviews, error) {
	rows, err := db.Queryx(`
		SELECT r.id, r.cid, r.usn, r.ease, r.time, r.type,
			CAST(r.factor AS real)/1000 AS factor,
			CASE
				WHEN r.ivl < 0 THEN -ivl
				ELSE r.ivl*24*60*60
			END AS ivl,
			CASE
				WHEN r.lastIvl < 0 THEN -ivl
				ELSE r.lastIvl*24*60*60
			END AS lastIvl
		FROM revlog r
		LEFT JOIN graves g ON g.oid=r.cid AND g.type=0
		WHERE g.oid IS NULL
		ORDER BY id DESc
	`)
	return &Reviews{rows}, err
}

func (db *DB) Close() (e error) {
	if db.tmpFile != "" {
		if err := os.Remove(db.tmpFile); err != nil {
			fmt.Printf("Cannot remove file: %s", err)
			e = err
		}
	}
	if db.DB != nil {
		if err := db.DB.Close(); err != nil {
			e = err
		}
	}
	return
}

func OpenOriginalDB(filename string) (db *DB, e error) {
	db = &DB{}
	sqldb, err := sqlx.Connect("sqlite3", filename)
	if err != nil {
		return db, err
	}
	db.DB = sqldb
	return db, nil
}

func OpenDB(src io.Reader) (db *DB, e error) {
	db = &DB{}
	dbFile, err := dumpToTemp(src)
	db.tmpFile = dbFile
	if err != nil {
		return db, err
	}
	sqldb, err := sqlx.Connect("sqlite3", dbFile)
	if err != nil {
		return db, err
	}
	db.DB = sqldb
	return db, nil
}

func dumpToTemp(src io.Reader) (string, error) {
	tmp, err := ioutil.TempFile("/tmp", "anki-sqlite3-")
	if err != nil {
		return "", err
	}
	defer tmp.Close()
	if _, err := io.Copy(tmp, src); err != nil {
		return "", err
	}
	return tmp.Name(), nil
}
