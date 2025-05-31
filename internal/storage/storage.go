// Package storage provides persistence functionality for pack allocation results.
// It defines interfaces and implementations for storing and retrieving allocation data.
//
// The package currently implements SQLite-based storage, but the interface
// allows for other storage backends to be implemented.
package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Common errors
var (
	ErrInvalidArgument = errors.New("invalid argument")
)

// Allocation represents a stored pack allocation result.
// It contains the order quantity, the calculated pack distribution,
// the total number of items, and when the allocation was created.
type Allocation struct {
	ID            int64
	OrderQuantity int
	Packs         map[int]int
	Total         int
	CreatedAt     time.Time
}

// Storage defines the interface for persistence operations.
// Implementations should provide thread-safe storage and retrieval
// of pack allocation results.
type Storage interface {
	// StoreAllocation saves a pack allocation result.
	// Returns an error if the operation fails or if the input is invalid.
	StoreAllocation(quantity int, packs map[int]int, total int) error

	// GetRecentAllocations retrieves the most recent allocations.
	// The limit parameter controls how many allocations to return.
	// Returns an error if the operation fails.
	GetRecentAllocations(limit int) ([]Allocation, error)

	// GetAllocationByQuantity retrieves the most recent allocation for a given quantity.
	// Returns nil if no allocation is found for the quantity.
	// Returns an error if the operation fails.
	GetAllocationByQuantity(quantity int) (*Allocation, error)

	// Close closes the storage connection.
	// It should be called when the storage is no longer needed.
	Close() error
}

// SQLiteStorage implements Storage using SQLite.
// It provides persistent storage of allocation results in a SQLite database.
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage instance.
// The dbPath parameter specifies the path to the SQLite database file.
// If the database doesn't exist, it will be created with the necessary schema.
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// Create tables if they don't exist
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS allocations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			order_quantity INTEGER NOT NULL,
			packs TEXT NOT NULL,
			total INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX IF NOT EXISTS idx_order_quantity ON allocations(order_quantity);
		CREATE INDEX IF NOT EXISTS idx_created_at ON allocations(created_at);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return &SQLiteStorage{db: db}, nil
}

// StoreAllocation saves a pack allocation result to the SQLite database.
// The packs map is stored as a JSON string in the database.
// Returns an error if the operation fails or if packs is nil.
func (s *SQLiteStorage) StoreAllocation(quantity int, packs map[int]int, total int) error {
	if packs == nil {
		return ErrInvalidArgument
	}

	packsJSON, err := json.Marshal(packs)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(
		"INSERT INTO allocations (order_quantity, packs, total) VALUES (?, ?, ?)",
		quantity, string(packsJSON), total,
	)
	return err
}

// GetRecentAllocations retrieves the most recent allocations from the database.
// Results are ordered by creation time in descending order.
// The limit parameter controls how many allocations to return.
func (s *SQLiteStorage) GetRecentAllocations(limit int) ([]Allocation, error) {
	rows, err := s.db.Query(
		"SELECT id, order_quantity, packs, total, created_at FROM allocations ORDER BY created_at DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var allocations []Allocation
	for rows.Next() {
		var a Allocation
		var packsJSON string
		err := rows.Scan(&a.ID, &a.OrderQuantity, &packsJSON, &a.Total, &a.CreatedAt)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(packsJSON), &a.Packs)
		if err != nil {
			return nil, err
		}

		allocations = append(allocations, a)
	}

	return allocations, rows.Err()
}

// GetAllocationByQuantity retrieves the most recent allocation for a given quantity.
// Returns nil if no allocation is found for the quantity.
func (s *SQLiteStorage) GetAllocationByQuantity(quantity int) (*Allocation, error) {
	var a Allocation
	var packsJSON string
	err := s.db.QueryRow(
		"SELECT id, order_quantity, packs, total, created_at FROM allocations WHERE order_quantity = ? ORDER BY created_at DESC LIMIT 1",
		quantity,
	).Scan(&a.ID, &a.OrderQuantity, &packsJSON, &a.Total, &a.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(packsJSON), &a.Packs)
	if err != nil {
		return nil, err
	}

	return &a, nil
}

// Close closes the SQLite database connection.
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
