package storage

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func setupTestDB(t *testing.T) (*SQLiteStorage, func()) {
	// Create a temporary database file
	dbPath := "test.db"
	storage, err := NewSQLiteStorage(dbPath)
	assert.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		storage.Close()
		os.Remove(dbPath)
	}

	return storage, cleanup
}

func TestStoreAndGetAllocation(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Test data
	quantity := 50
	packs := map[int]int{23: 1, 31: 1}
	total := 54

	// Store allocation
	err := storage.StoreAllocation(quantity, packs, total)
	assert.NoError(t, err)

	// Retrieve allocation
	allocation, err := storage.GetAllocationByQuantity(quantity)
	assert.NoError(t, err)
	assert.NotNil(t, allocation)
	assert.Equal(t, quantity, allocation.OrderQuantity)
	assert.Equal(t, packs, allocation.Packs)
	assert.Equal(t, total, allocation.Total)
	assert.True(t, time.Since(allocation.CreatedAt) < time.Minute)
}

func TestGetRecentAllocations(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Store multiple allocations
	allocations := []struct {
		quantity int
		packs    map[int]int
		total    int
	}{
		{50, map[int]int{23: 1, 31: 1}, 54},
		{100, map[int]int{53: 1, 31: 1, 23: 1}, 107},
		{200, map[int]int{53: 3, 31: 1, 23: 1}, 213},
	}

	for _, a := range allocations {
		err := storage.StoreAllocation(a.quantity, a.packs, a.total)
		assert.NoError(t, err)
	}

	// Test getting all allocations
	recent, err := storage.GetRecentAllocations(10)
	assert.NoError(t, err)
	assert.Len(t, recent, 3)

	// Verify order (most recent first)
	assert.Equal(t, 200, recent[0].OrderQuantity)
	assert.Equal(t, 100, recent[1].OrderQuantity)
	assert.Equal(t, 50, recent[2].OrderQuantity)

	// Test limit
	recent, err = storage.GetRecentAllocations(2)
	assert.NoError(t, err)
	assert.Len(t, recent, 2)
	assert.Equal(t, 200, recent[0].OrderQuantity)
	assert.Equal(t, 100, recent[1].OrderQuantity)
}

func TestGetAllocationByQuantity(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Test non-existent quantity
	allocation, err := storage.GetAllocationByQuantity(999)
	assert.NoError(t, err)
	assert.Nil(t, allocation)

	// Store and retrieve allocation
	quantity := 50
	packs := map[int]int{23: 1, 31: 1}
	total := 54

	err = storage.StoreAllocation(quantity, packs, total)
	assert.NoError(t, err)

	allocation, err = storage.GetAllocationByQuantity(quantity)
	assert.NoError(t, err)
	assert.NotNil(t, allocation)
	assert.Equal(t, quantity, allocation.OrderQuantity)
	assert.Equal(t, packs, allocation.Packs)
	assert.Equal(t, total, allocation.Total)
}

func TestStoreAllocationWithInvalidData(t *testing.T) {
	storage, cleanup := setupTestDB(t)
	defer cleanup()

	// Test with nil packs
	err := storage.StoreAllocation(50, nil, 50)
	assert.ErrorIs(t, err, ErrInvalidArgument)

	// Test with empty packs
	err = storage.StoreAllocation(50, map[int]int{}, 50)
	assert.NoError(t, err)
}
