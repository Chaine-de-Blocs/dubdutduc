package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var ErrNotFound = errors.New("block not found")

type Block struct {
	Height int64  `json:"height"`
	Hash   string `json:"hash"`
	Time   int64  `json:"time"`
	Tonic  string `json:"tonic"`
	Mode   string `json:"mode"`
	BPM    int    `json:"bpm"`
	Beats  int    `json:"beats"`
}

type Store struct {
	db *sql.DB
}

const schema = `
CREATE TABLE IF NOT EXISTS blocks (
	height INTEGER PRIMARY KEY,
	hash   TEXT    NOT NULL,
	time   INTEGER NOT NULL,
	tonic  TEXT    NOT NULL,
	mode   TEXT    NOT NULL,
	bpm    INTEGER NOT NULL,
	beats  INTEGER NOT NULL
)`

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

// Save inserts the block, replacing a block of the same height only when the chain reorganized.
func (s *Store) Save(ctx context.Context, b Block) (bool, error) {
	var prev string
	err := s.db.QueryRowContext(ctx, `SELECT hash FROM blocks WHERE height = ?`, b.Height).Scan(&prev)
	isNew := errors.Is(err, sql.ErrNoRows)
	if err != nil && !isNew {
		return false, err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO blocks (height, hash, time, tonic, mode, bpm, beats)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(height) DO UPDATE SET
			hash = excluded.hash, time = excluded.time, tonic = excluded.tonic,
			mode = excluded.mode, bpm = excluded.bpm, beats = excluded.beats`,
		b.Height, b.Hash, b.Time, b.Tonic, b.Mode, b.BPM, b.Beats)
	return isNew || prev != b.Hash, err
}

func (s *Store) List(ctx context.Context) ([]Block, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT height, hash, time, tonic, mode, bpm, beats FROM blocks ORDER BY height DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks := []Block{}
	for rows.Next() {
		var b Block
		if err := rows.Scan(&b.Height, &b.Hash, &b.Time, &b.Tonic, &b.Mode, &b.BPM, &b.Beats); err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
	}
	return blocks, rows.Err()
}

func (s *Store) Get(ctx context.Context, height int64) (Block, error) {
	var b Block
	err := s.db.QueryRowContext(ctx,
		`SELECT height, hash, time, tonic, mode, bpm, beats FROM blocks WHERE height = ?`, height).
		Scan(&b.Height, &b.Hash, &b.Time, &b.Tonic, &b.Mode, &b.BPM, &b.Beats)
	if errors.Is(err, sql.ErrNoRows) {
		return b, ErrNotFound
	}
	return b, err
}

// MaxHeight returns the highest stored height, or false when the catalog is empty.
func (s *Store) MaxHeight(ctx context.Context) (int64, bool, error) {
	var h sql.NullInt64
	if err := s.db.QueryRowContext(ctx, `SELECT MAX(height) FROM blocks`).Scan(&h); err != nil {
		return 0, false, err
	}
	return h.Int64, h.Valid, nil
}

// Heights returns the stored heights in [lo, hi].
func (s *Store) Heights(ctx context.Context, lo, hi int64) (map[int64]bool, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT height FROM blocks WHERE height BETWEEN ? AND ?`, lo, hi)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	have := map[int64]bool{}
	for rows.Next() {
		var h int64
		if err := rows.Scan(&h); err != nil {
			return nil, err
		}
		have[h] = true
	}
	return have, rows.Err()
}
