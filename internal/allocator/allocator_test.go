package allocator

import (
	"testing"

	"github.com/n-th/gymshark/internal/storage"
	"github.com/stretchr/testify/assert"
)

// mockStorage implements storage.Storage for testing
type mockStorage struct {
	allocations map[int]*storage.Allocation
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		allocations: make(map[int]*storage.Allocation),
	}
}

func (m *mockStorage) StoreAllocation(quantity int, packs map[int]int, total int) error {
	m.allocations[quantity] = &storage.Allocation{
		OrderQuantity: quantity,
		Packs:         packs,
		Total:         total,
	}
	return nil
}

func (m *mockStorage) GetRecentAllocations(limit int) ([]storage.Allocation, error) {
	// Not used in tests
	return nil, nil
}

func (m *mockStorage) GetAllocationByQuantity(quantity int) (*storage.Allocation, error) {
	return m.allocations[quantity], nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestCalculatePacks(t *testing.T) {
	storage := newMockStorage()
	allocator := NewAllocator([]int{23, 31, 53}, storage)

	tests := []struct {
		name           string
		quantity       int
		expectedPacks  map[int]int
		expectedTotal  int
		expectedError  bool
		errorMessage   string
		shouldBeCached bool
	}{
		{
			name:          "valid quantity",
			quantity:      50,
			expectedPacks: map[int]int{53: 1},
			expectedTotal: 53,
			expectedError: false,
		},
		{
			name:          "zero quantity",
			quantity:      0,
			expectedPacks: nil,
			expectedTotal: 0,
			expectedError: true,
			errorMessage:  "quantity must be greater than 0",
		},
		{
			name:          "negative quantity",
			quantity:      -10,
			expectedPacks: nil,
			expectedTotal: 0,
			expectedError: true,
			errorMessage:  "quantity must be greater than 0",
		},
		{
			name:          "small quantity",
			quantity:      10,
			expectedPacks: map[int]int{23: 1},
			expectedTotal: 23,
			expectedError: false,
		},
		{
			name:          "large quantity",
			quantity:      500,
			expectedPacks: map[int]int{53: 9, 23: 1},
			expectedTotal: 500,
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packs, total, err := allocator.CalculatePacks(tt.quantity)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMessage, err.Error())
				assert.Nil(t, packs)
				assert.Equal(t, 0, total)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedPacks, packs)
				assert.Equal(t, tt.expectedTotal, total)

				// Verify the result was stored
				cached, err := storage.GetAllocationByQuantity(tt.quantity)
				assert.NoError(t, err)
				assert.NotNil(t, cached)
				assert.Equal(t, tt.quantity, cached.OrderQuantity)
				assert.Equal(t, tt.expectedPacks, cached.Packs)
				assert.Equal(t, tt.expectedTotal, cached.Total)
			}
		})
	}
}

func TestCalculatePacksWithDifferentPackSizes(t *testing.T) {
	storage := newMockStorage()
	allocator := NewAllocator([]int{5, 10, 20}, storage)

	tests := []struct {
		name          string
		quantity      int
		expectedPacks map[int]int
		expectedTotal int
	}{
		{
			name:          "small quantity",
			quantity:      7,
			expectedPacks: map[int]int{10: 1},
			expectedTotal: 10,
		},
		{
			name:          "medium quantity",
			quantity:      25,
			expectedPacks: map[int]int{20: 1, 5: 1},
			expectedTotal: 25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packs, total, err := allocator.CalculatePacks(tt.quantity)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedPacks, packs)
			assert.Equal(t, tt.expectedTotal, total)

			// Verify the result was stored
			cached, err := storage.GetAllocationByQuantity(tt.quantity)
			assert.NoError(t, err)
			assert.NotNil(t, cached)
			assert.Equal(t, tt.quantity, cached.OrderQuantity)
			assert.Equal(t, tt.expectedPacks, cached.Packs)
			assert.Equal(t, tt.expectedTotal, cached.Total)
		})
	}
}

func TestNewAllocator(t *testing.T) {
	tests := []struct {
		name           string
		packSizes      []int
		expectedSorted []int
	}{
		{
			name:           "unsorted pack sizes",
			packSizes:      []int{31, 23, 53},
			expectedSorted: []int{53, 31, 23},
		},
		{
			name:           "already sorted pack sizes",
			packSizes:      []int{53, 31, 23},
			expectedSorted: []int{53, 31, 23},
		},
		{
			name:           "single pack size",
			packSizes:      []int{23},
			expectedSorted: []int{23},
		},
		{
			name:           "empty pack sizes",
			packSizes:      []int{},
			expectedSorted: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := newMockStorage()
			allocator := NewAllocator(tt.packSizes, storage)
			assert.Equal(t, tt.expectedSorted, allocator.packSizes)
		})
	}
}

func TestGetRecentAllocations(t *testing.T) {
	storage := newMockStorage()
	allocator := NewAllocator([]int{23, 31, 53}, storage)

	// Test when storage is not configured
	allocator.storage = nil
	allocations, err := allocator.GetRecentAllocations(10)
	assert.Error(t, err)
	assert.Equal(t, "storage not configured", err.Error())
	assert.Nil(t, allocations)

	// Test with mock storage
	allocator.storage = storage
	allocations, err = allocator.GetRecentAllocations(10)
	assert.NoError(t, err)
	assert.Empty(t, allocations)
}
