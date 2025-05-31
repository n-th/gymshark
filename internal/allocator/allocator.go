// Package allocator provides functionality for calculating optimal pack distributions
// for fulfilling orders with fixed pack sizes.
//
// The package implements a smart algorithm that determines the most efficient
// combination of pack sizes to fulfill an order quantity while minimizing waste.
// It also includes persistence capabilities for storing and retrieving allocation results.
package allocator

import (
	"errors"
	"log"
	"sort"

	"github.com/n-th/gymshark/internal/storage"
)

// Common errors
var (
	ErrInvalidQuantity      = errors.New("quantity must be greater than 0")
	ErrStorageNotConfigured = errors.New("storage not configured")
)

// Pack represents a pack size and its quantity
type Pack struct {
	Size     int
	Quantity int
}

// Allocator handles pack size calculations and result persistence.
// It maintains a sorted list of available pack sizes and can store
// calculation results in a persistent storage.
type Allocator struct {
	packSizes []int
	storage   storage.Storage
}

// NewAllocator creates a new Allocator instance with the specified pack sizes.
// The pack sizes are sorted in ascending order for optimal calculation.
// If storage is provided, calculation results will be persisted.
func NewAllocator(packSizes []int, storage storage.Storage) *Allocator {
	// Create a copy of pack sizes to avoid modifying the input slice
	sizes := make([]int, len(packSizes))
	copy(sizes, packSizes)

	// Sort pack sizes in ascending order
	sort.Ints(sizes)

	return &Allocator{
		packSizes: sizes,
		storage:   storage,
	}
}

// CalculatePacks determines the optimal distribution of packs for a given quantity.
// It returns a map of pack sizes to quantities, the total number of items,
// and any error that occurred during calculation.
//
// The algorithm works as follows:
// 1. For each possible combination of pack sizes, calculate the total items
// 2. Choose the combination that minimizes waste while meeting or exceeding the order quantity
// 3. If multiple combinations have the same waste, prefer the one with fewer total packs
//
// Example:
//
//	alloc := NewAllocator([]int{23, 31, 53}, nil)
//	packs, total, err := alloc.CalculatePacks(50)
//	// packs: map[23:1 31:1]
//	// total: 54
func (a *Allocator) CalculatePacks(quantity int) (map[int]int, int, error) {
	if quantity <= 0 {
		return nil, 0, ErrInvalidQuantity
	}

	// If we have storage, check for cached result
	if a.storage != nil {
		if cached, err := a.storage.GetAllocationByQuantity(quantity); err == nil && cached != nil {
			log.Printf("Using cached result for quantity %d", quantity)
			return cached.Packs, cached.Total, nil
		}
	}

	// Sort pack sizes in descending order for optimal calculation
	sort.Sort(sort.Reverse(sort.IntSlice(a.packSizes)))

	// Initialize variables to track the best solution
	bestPacks := make(map[int]int)
	bestTotal := 0
	bestWaste := -1
	bestPackCount := -1

	// Try different combinations of pack sizes
	for i := 0; i < len(a.packSizes); i++ {
		remaining := quantity
		currentPacks := make(map[int]int)
		currentTotal := 0

		// Start with the current pack size
		for j := i; j < len(a.packSizes); j++ {
			size := a.packSizes[j]
			if remaining <= 0 {
				break
			}

			// Calculate how many packs of this size we need
			numPacks := (remaining + size - 1) / size
			currentPacks[size] = numPacks
			currentTotal += size * numPacks
			remaining -= size * numPacks
		}

		// Calculate waste (items over the order quantity)
		waste := currentTotal - quantity
		packCount := 0
		for _, count := range currentPacks {
			packCount += count
		}

		// Update best solution if:
		// 1. This is the first valid solution, or
		// 2. This solution has less waste, or
		// 3. This solution has the same waste but fewer total packs
		if bestWaste == -1 || waste < bestWaste || (waste == bestWaste && packCount < bestPackCount) {
			bestPacks = currentPacks
			bestTotal = currentTotal
			bestWaste = waste
			bestPackCount = packCount
		}
	}

	// Store the result if we have storage
	if a.storage != nil {
		if err := a.storage.StoreAllocation(quantity, bestPacks, bestTotal); err != nil {
			log.Printf("Failed to store allocation: %v", err)
		}
	}

	return bestPacks, bestTotal, nil
}

// findMinimumItems finds the minimum number of items needed to fulfill the order
func (a *Allocator) findMinimumItems(orderQuantity int) int {
	smallestPack := a.packSizes[len(a.packSizes)-1]

	baseQuantity := (orderQuantity + smallestPack - 1) / smallestPack
	baseItems := baseQuantity * smallestPack

	for _, size := range a.packSizes {
		if size <= orderQuantity {
			continue
		}

		if size < baseItems {
			log.Printf("Using single pack of size %d for order %d (less overage)", size, orderQuantity)
			return size
		}
	}

	return baseItems
}

// GetRecentAllocations retrieves the most recent pack allocations from storage.
// It returns up to the specified limit of allocations, ordered by creation time.
// Returns an error if storage is not configured or if the retrieval fails.
func (a *Allocator) GetRecentAllocations(limit int) ([]storage.Allocation, error) {
	if a.storage == nil {
		return nil, ErrStorageNotConfigured
	}
	return a.storage.GetRecentAllocations(limit)
}

// Close closes the allocator's storage connection if one is configured.
// It should be called when the allocator is no longer needed.
func (a *Allocator) Close() error {
	if a.storage != nil {
		return a.storage.Close()
	}
	return nil
}
