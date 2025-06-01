package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	return []storage.Allocation{}, nil
}

func (m *mockStorage) GetAllocationByQuantity(quantity int) (*storage.Allocation, error) {
	return m.allocations[quantity], nil
}

func (m *mockStorage) Close() error {
	return nil
}

func setupTestRouter() (*gin.Engine, *Handler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	storage := newMockStorage()
	alloc := allocator.NewAllocator([]int{23, 31, 53}, storage)
	handler := NewHandler(alloc)
	handler.RegisterRoutes(router)
	return router, handler
}

func TestCalculatePacks(t *testing.T) {
	router, _ := setupTestRouter()

	tests := []struct {
		name           string
		quantity       string
		expectedStatus int
		expectedBody   map[string]interface{}
		expectedError  string
	}{
		{
			name:           "valid quantity 50",
			quantity:       "50",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"packs": map[string]interface{}{
					"23": float64(1),
					"31": float64(1),
				},
				"total": float64(54),
			},
		},
		{
			name:           "valid quantity 10",
			quantity:       "10",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"packs": map[string]interface{}{
					"23": float64(1),
				},
				"total": float64(23),
			},
		},
		{
			name:           "valid quantity 500",
			quantity:       "500",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"packs": map[string]interface{}{
					"53": float64(9),
					"23": float64(1),
				},
				"total": float64(500),
			},
		},
		{
			name:           "missing quantity",
			quantity:       "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid quantity: strconv.Atoi: parsing \"\": invalid syntax",
		},
		{
			name:           "invalid quantity",
			quantity:       "abc",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "invalid quantity: strconv.Atoi: parsing \"abc\": invalid syntax",
		},
		{
			name:           "zero quantity",
			quantity:       "0",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "quantity must be greater than 0",
		},
		{
			name:           "negative quantity",
			quantity:       "-10",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "quantity must be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request to the calculate endpoint
			req := httptest.NewRequest("GET", "/calculate?quantity="+tt.quantity, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check the status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse the response body
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectedError != "" {
				// Check error response
				assert.Equal(t, tt.expectedError, response["error"])
			} else {
				// Check successful response
				assert.Equal(t, tt.expectedBody["packs"], response["packs"])
				assert.Equal(t, tt.expectedBody["total"], response["total"])
			}
		})
	}
}

func TestHealthCheck(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestGetRecentAllocations(t *testing.T) {
	router, _ := setupTestRouter()

	req := httptest.NewRequest("GET", "/recent", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["allocations"])
}

func TestCORSHeaders(t *testing.T) {
	router, _ := setupTestRouter()

	// Test preflight request
	req := httptest.NewRequest("OPTIONS", "/calculate", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))

	// Test actual request
	req = httptest.NewRequest("GET", "/calculate?quantity=50", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type", w.Header().Get("Access-Control-Allow-Headers"))
}
