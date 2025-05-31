package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/n-th/gymshark/internal/allocator"
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
	return nil, nil
}

func (m *mockStorage) GetAllocationByQuantity(quantity int) (*storage.Allocation, error) {
	return m.allocations[quantity], nil
}

func (m *mockStorage) Close() error {
	return nil
}

func TestCalculatePacks(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new Gin router
	router := gin.New()

	// Create a new allocator with mock storage
	storage := newMockStorage()
	alloc := allocator.NewAllocator([]int{23, 31, 53}, storage)

	// Create a new handler
	handler := NewHandler(alloc)

	// Register the routes
	handler.RegisterRoutes(router)

	tests := []struct {
		name           string
		quantity       string
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:           "valid quantity",
			quantity:       "50",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"packs": map[string]interface{}{
					"23": float64(2),
				},
			},
		},
		{
			name:           "invalid quantity - non-numeric",
			quantity:       "abc",
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "invalid quantity: strconv.Atoi: parsing \"abc\": invalid syntax",
			},
		},
		{
			name:           "invalid quantity - zero",
			quantity:       "0",
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "quantity must be greater than 0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new request
			req, err := http.NewRequest("GET", "/calculate?quantity="+tt.quantity, nil)
			assert.NoError(t, err)

			// Create a response recorder
			w := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(w, req)

			// Check the status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse the response body
			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			// Check the response body
			assert.Equal(t, tt.expectedBody, response)

			// If the request was successful, verify the result was stored
			if tt.expectedStatus == http.StatusOK {
				quantity, _ := strconv.Atoi(tt.quantity)
				cached, err := storage.GetAllocationByQuantity(quantity)
				assert.NoError(t, err)
				assert.NotNil(t, cached)
				assert.Equal(t, quantity, cached.OrderQuantity)
				assert.Equal(t, tt.expectedBody["packs"], cached.Packs)
			}
		})
	}
}

func TestGetRecentAllocations(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new Gin router
	router := gin.New()

	// Test when storage is not configured
	alloc := allocator.NewAllocator([]int{23, 31, 53}, nil)
	handler := NewHandler(alloc)
	handler.RegisterRoutes(router)

	req, err := http.NewRequest("GET", "/recent", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "storage not configured", response["error"])

	// Create a new router for the next test
	router = gin.New()

	// Test with mock storage
	storage := newMockStorage()
	alloc = allocator.NewAllocator([]int{23, 31, 53}, storage)
	handler = NewHandler(alloc)
	handler.RegisterRoutes(router)

	req, err = http.NewRequest("GET", "/recent", nil)
	assert.NoError(t, err)

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Empty(t, response["allocations"])
}
