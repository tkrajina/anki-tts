// Copyright: Jonathan Hall
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html

package anki

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"io"

	"github.com/jmoiron/sqlx"
)

// Apkg manages state of an Anki package file during processing.
type Apkg struct {
	reader *zip.Reader
	closer *zip.ReadCloser
	sqlite *zip.File
	media  *zipIndex
	db     *DB
}

// ReadFile reads an *.apkg file, returning an Apkg struct for processing.
func ReadFile(f string) (*Apkg, error) {
	z, err := zip.OpenReader(f)
	if err != nil {
		return nil, err
	}
	a := &Apkg{
		reader: &z.Reader,
		closer: z,
	}
	return a, a.open()
}

// ReadBytes reads an *.apkg file from a bytestring, returning an Apkg struct
// for processing.
func ReadBytes(b []byte) (*Apkg, error) {
	r := bytes.NewReader(b)
	return ReadReader(r, int64(len(b)))
}

// ReadReader reads an *.apkg file from an io.Reader, returning an Apkg struct
// for processing.
func ReadReader(r io.ReaderAt, size int64) (*Apkg, error) {
	z, err := zip.NewReader(r, size)
	if err != nil {
		return nil, err
	}
	a := &Apkg{
		reader: z,
	}
	return a, a.open()
}

func (a *Apkg) open() error {
	if err := a.populateIndex(); err != nil {
		return err
	}
	rc, err := a.sqlite.Open()
	if err != nil {
		return err
	}
	defer rc.Close()
	db, err := OpenDB(rc)
	if err != nil {
		return err
	}
	a.db = db
	return nil
}

type zipIndex struct {
	index map[string]*zip.File
}

func (a *Apkg) ReadMediaFile(name string) ([]byte, error) {
	return a.media.ReadFile(name)
}

func (zi *zipIndex) ReadFile(name string) ([]byte, error) {
	zipFile, ok := zi.index[name]
	if !ok {
		return nil, errors.New("File `" + name + "` not found in zip index")
	}
	fh, err := zipFile.Open()
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(fh)
	return buf.Bytes(), nil
}

func (a *Apkg) populateIndex() error {
	index := &zipIndex{
		index: make(map[string]*zip.File),
	}
	for _, file := range a.reader.File {
		index.index[file.FileHeader.Name] = file
	}

	if sqlite, ok := index.index["collection.anki2"]; !ok {
		return errors.New("Unable to find `collection.anki2` in archive")
	} else {
		a.sqlite = sqlite
	}

	mediaFile, err := index.ReadFile("media")
	if err != nil {
		return err
	}

	mediaMap := make(map[string]string)
	if err := json.Unmarshal(mediaFile, &mediaMap); err != nil {
		return err
	}
	a.media = &zipIndex{
		index: make(map[string]*zip.File),
	}
	for idx, filename := range mediaMap {
		a.media.index[filename] = index.index[idx]
	}
	return nil
}

// Close closes any opened resources (io.Reader, SQLite handles, etc). Any
// subsequent calls to extant objects (Collection, Decks, Notes, etc) which
// depend on these resources may fail. Only call this method after you're
// completely done reading the Apkg file.
func (a *Apkg) Close() (e error) {
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			e = err
		}
	}
	if a.closer != nil {
		if err := a.closer.Close(); err != nil {
			e = err
		}
	}
	return
}

func (a *Apkg) Collection() (*Collection, error) {
	return a.db.Collection()
}

// Notes is a wrapper around sqlx.Rows, which means that any standard sqlx.Rows
// or sql.Rows methods may be called on it. Generally, you should only ever
// need to call Next() and Close(), in addition to Note() which is defined in
// this package.
type Notes struct {
	*sqlx.Rows
}

// Notes returns a Notes struct representing all of the Note rows in the *.apkg
// package file.
func (a *Apkg) Notes() (*Notes, error) {
	return a.db.Notes()
}

// Note is a simple wrapper around sqlx's StructScan(), which returns a Note
// struct populated from the database.
func (n *Notes) Note() (*Note, error) {
	note := &Note{}
	err := n.StructScan(note)
	return note, err
}

// Cards is a wrapper around sqlx.Rows, which means that any standard sqlx.Rows
// or sql.Rows methods may be called on it. Generally, you should only ever
// need to call Next() and Close(), in addition to Card() which is defined in
// this package.
type Cards struct {
	*sqlx.Rows
}

// Cards returns a Cards struct represeting all of the non-deleted cards in the
// *.apkg package file.
func (a *Apkg) Cards() (*Cards, error) {
	return a.db.Cards()
}

func (c *Cards) Card() (*Card, error) {
	card := &Card{}
	err := c.StructScan(card)
	return card, err
}

type Reviews struct {
	*sqlx.Rows
}

// Reviews returns a Reviews struct representing all of the reviews of
// non-deleted cards in the *.apkg package file, in reverse chronological
// order (newest first).
func (a *Apkg) Reviews() (*Reviews, error) {
	return a.db.Reviews()
}

func (r *Reviews) Review() (*Review, error) {
	review := &Review{}
	err := r.StructScan(review)
	return review, err
}
