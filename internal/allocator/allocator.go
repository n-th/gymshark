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
	sizes := make([]int, len(packSizes))
	copy(sizes, packSizes)
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
	return &Allocator{packSizes: sizes, storage: s}
}

// CalculatePacksOptimized calculates the optimal pack distribution for a given quantity
// using the stored pack sizes.
// It returns the pack distribution, the total quantity, and an error if the quantity is invalid.
func (a *Allocator) CalculatePacksOptimized(quantity int) (map[int]int, int, error) {

	if quantity <= 0 {
		return nil, 0, ErrInvalidQuantity
	}

	if a.storage != nil {
		if cached, err := a.storage.GetAllocationByQuantity(quantity); err == nil && cached != nil {
			log.Printf("Using cached result for quantity %d", quantity)
			return cached.Packs, cached.Total, nil
		}
	}

	maxPackSize := a.packSizes[0]
	maxPossibleWaste := maxPackSize * 1000000 // arbitrarily large upper bound

	best := struct {
		packs     map[int]int
		total     int
		waste     int
		packCount int
		found     bool
	}{
		waste: maxPossibleWaste,
	}

	a.findOptimal(quantity, 0, map[int]int{}, 0, 0, &best)

	if !best.found {
		return nil, 0, errors.New("no valid pack combination found")
	}

	if a.storage != nil {
		if err := a.storage.StoreAllocation(quantity, best.packs, best.total); err != nil {
			log.Printf("Failed to store allocation: %v", err)

		}
	}

	return best.packs, best.total, nil
}

// findOptimal is a helper function that finds the optimal pack distribution
// for a given quantity using a recursive backtracking approach.
func (a *Allocator) findOptimal(target, index int, current map[int]int, total, packCount int, best *struct {
	packs     map[int]int
	total     int
	waste     int
	packCount int
	found     bool
}) {
	if total >= target {
		waste := total - target
		if !best.found || waste < best.waste || (waste == best.waste && packCount < best.packCount) {
			best.found = true
			best.total = total
			best.waste = waste
			best.packCount = packCount
			best.packs = cloneMap(current)
		}
		return
	}

	if index >= len(a.packSizes) {
		return

	}

	size := a.packSizes[index]
	maxQty := (target - total + size - 1) / size // minimal fill

	for q := maxQty; q >= 0; q-- {
		if q > 0 {
			current[size] = q
		} else {
			delete(current, size)
		}
		a.findOptimal(target, index+1, current, total+q*size, packCount+q, best)
	}
}

// GreedyWithCorrectionPacks computes an approximate pack distribution
// using a greedy approach followed by local correction to reduce waste.
func (a *Allocator) GreedyWithCorrectionPacks(quantity int) (map[int]int, int) {
	packSizes := a.packSizes
	packs := make(map[int]int)
	total := 0
	remaining := quantity

	// Greedy phase: use as many large packs as possible
	for _, size := range packSizes {
		count := remaining / size
		if count > 0 {
			packs[size] = count
			total += size * count
			remaining -= size * count
		}
	}

	// Add smallest pack if needed
	if remaining > 0 {
		smallest := packSizes[len(packSizes)-1]
		packs[smallest]++
		total += smallest
	}

	// Local correction phase
	// Try replacing a small pack with a larger one that reduces waste
	for i := len(packSizes) - 1; i > 0; i-- {
		small := packSizes[i]
		if packs[small] == 0 {
			continue
		}
		for j := i - 1; j >= 0; j-- {
			large := packSizes[j]
			newTotal := total - small + large
			if newTotal >= quantity && newTotal < total {
				packs[small]--
				if packs[small] == 0 {
					delete(packs, small)
				}
				packs[large]++
				total = newTotal
				break
			}
		}
	}

	return packs, total
}

// cloneMap creates a deep copy of a map[int]int.
func cloneMap(src map[int]int) map[int]int {
	dst := make(map[int]int, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
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

	// Special case: order is smaller than all pack sizes
	smallest := a.packSizes[len(a.packSizes)-1]
	if orderQuantity < smallest {
		result := map[int]int{smallest: 1}
		if a.storage != nil {
			if err := a.storage.StoreAllocation(orderQuantity, result, smallest); err != nil {
				log.Printf("Failed to store allocation: %v", err)
			}
		}
		return result, smallest, nil
	}

	// Initialize result map
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

	if a.storage != nil {
		if err := a.storage.StoreAllocation(orderQuantity, result, bestTotal); err != nil {
			log.Printf("Failed to store allocation: %v", err)
		}
	}

	return result, bestTotal, nil
}
