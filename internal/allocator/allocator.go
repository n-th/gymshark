// Package allocator provides optimal pack distribution calculation
// to fulfill order quantities using fixed pack sizes.
package allocator

import (
	"errors"
	"log"
	"sort"

	"github.com/n-th/gymshark/internal/storage"
)

var (
	ErrInvalidQuantity      = errors.New("quantity must be greater than 0")
	ErrStorageNotConfigured = errors.New("storage not configured")
)

type Pack struct {
	Size     int
	Quantity int
}

type Allocator struct {
	packSizes []int
	storage   storage.Storage
}

func NewAllocator(packSizes []int, s storage.Storage) *Allocator {
	if packSizes == nil {
		packSizes = []int{}
	}
	sizes := append([]int(nil), packSizes...)
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
	return &Allocator{packSizes: sizes, storage: s}
}

// GetRecentAllocations retrieves the most recent allocations from the storage.
func (a *Allocator) GetRecentAllocations(limit int) ([]storage.Allocation, error) {
	if a.storage == nil {
		return nil, ErrStorageNotConfigured
	}
	return a.storage.GetRecentAllocations(limit)
}

// Close closes the storage.
func (a *Allocator) Close() error {
	if a.storage != nil {
		return a.storage.Close()
	}
	return nil
}

// CalculatePacks calculates the optimal pack distribution for a given order quantity
func (a *Allocator) CalculatePacks(orderQuantity int) (map[int]int, int, error) {
	log.Printf("Calculating optimal packs for order quantity: %d", orderQuantity)
	if orderQuantity <= 0 {
		log.Printf("Order quantity <= 0, returning error")
		return nil, 0, errors.New("quantity must be greater than 0")
	}

	if len(a.packSizes) == 0 {
		log.Printf("No pack sizes configured")
		return nil, 0, errors.New("no pack sizes configured")
	}

	result := make(map[int]int)

	// Calculate the minimum number of packs needed for each pack size
	minPacks := make(map[int]int)
	for _, size := range a.packSizes {
		minPacks[size] = (orderQuantity + size - 1) / size
	}

	// Find the best combination by trying different combinations
	bestResult := make(map[int]int)
	bestTotal := 0
	bestOverage := orderQuantity

	// Try combinations starting from the largest pack size
	for _, size := range a.packSizes {
		currentResult := make(map[int]int)
		currentTotal := 0
		remaining := orderQuantity

		// Start with the maximum possible number of current pack size
		maxPacks := minPacks[size]
		for packs := maxPacks; packs >= 0; packs-- {
			currentResult[size] = packs
			currentTotal = packs * size
			remaining = orderQuantity - currentTotal

			// If we've met or exceeded the order quantity, check if this is better
			if remaining <= 0 {
				overage := -remaining
				if overage < bestOverage || (overage == bestOverage && currentTotal < bestTotal) {
					bestResult = make(map[int]int)
					for k, v := range currentResult {
						bestResult[k] = v
					}
					bestTotal = currentTotal
					bestOverage = overage
				}
				continue
			}

			// Try to fill the remaining quantity with smaller packs
			for _, smallerSize := range a.packSizes {
				if smallerSize >= size {
					continue
				}
				smallerPacks := (remaining + smallerSize - 1) / smallerSize
				currentResult[smallerSize] = smallerPacks
				currentTotal += smallerPacks * smallerSize
				overage := currentTotal - orderQuantity

				if overage < bestOverage || (overage == bestOverage && currentTotal < bestTotal) {
					bestResult = make(map[int]int)
					for k, v := range currentResult {
						bestResult[k] = v
					}
					bestTotal = currentTotal
					bestOverage = overage
				}
			}
		}
	}

	// Copy the best result to the final result
	for k, v := range bestResult {
		if v > 0 {
			result[k] = v
		}
	}

	return result, bestTotal, nil
}
