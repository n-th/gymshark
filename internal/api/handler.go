// Package api provides HTTP handlers for the pack allocation service.
// It implements a RESTful API for calculating pack distributions
// and retrieving allocation history.
package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/n-th/gymshark/internal/allocator"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Handler handles HTTP requests for the pack allocation service.
// It provides endpoints for calculating pack distributions and
// retrieving allocation history.
type Handler struct {
	allocator *allocator.Allocator
}

// NewHandler creates a new handler instance.
// The allocator parameter is used for pack calculations and result persistence.
func NewHandler(allocator *allocator.Allocator) *Handler {
	return &Handler{
		allocator: allocator,
	}
}

// RegisterRoutes registers the API routes with the provided Gin router.
// The following endpoints are registered:
//   - GET /calculate - Calculate pack distribution for a quantity
//   - GET /recent - Get recent allocation history
//   - GET /health - Health check endpoint
//   - GET /swagger/*any - Swagger documentation
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.Status(http.StatusOK)
			c.Abort()
			return
		}
	})

	// API routes
	router.GET("/calculate", h.calculatePacks)
	router.GET("/recent", h.getRecentAllocations)

	// Health check
	router.GET("/health", h.healthCheck)

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// @Summary Calculate pack distribution
// @Description Calculate the optimal pack distribution for a given quantity
// @Tags packs
// @Accept json
// @Produce json
// @Param quantity query int true "Order quantity"
// @Success 200 {object} map[string]interface{} "Pack distribution"
// @Failure 400 {object} map[string]string "Error message"
// @Router /calculate [get]
func (h *Handler) calculatePacks(c *gin.Context) {
	quantityStr := c.Query("quantity")
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid quantity"})
		return
	}

	type allocationResult struct {
		Packs map[int]int
		Total int
		Err   error
	}

	resultChan := make(chan allocationResult, 1)
	ctx, cancel := context.WithTimeout(c.Request.Context(), 20*time.Minute)
	defer cancel()

	// const maxExactQuantity = 10000

	// if quantity >= maxExactQuantity {
	// 	packs, total := h.allocator.GreedyWithCorrectionPacks(quantity)
	// 	resultChan <- allocationResult{packs, total, err}
	// } else {
	// 	packs, total, err := h.allocator.CalculatePacksOptimized(quantity)
	// 	resultChan <- allocationResult{packs, total, err}
	// }

	packs, total, err := h.allocator.CalculatePacks(quantity)
	resultChan <- allocationResult{packs, total, err}

	select {
	case <-ctx.Done():
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": "calculation timeout"})
	case result := <-resultChan:
		if result.Err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": result.Err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"packs": result.Packs,
			"total": result.Total,
		})
	}
}

// @Summary Get recent allocations
// @Description Get the most recent pack allocations
// @Tags packs
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Recent allocations"
// @Failure 500 {object} map[string]string "Error message"
// @Router /recent [get]
func (h *Handler) getRecentAllocations(c *gin.Context) {
	allocations, err := h.allocator.GetRecentAllocations(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"allocations": allocations,
	})
}

// @Summary Health check
// @Description Check if the service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string "Health status"
// @Router /health [get]
func (h *Handler) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
