// Package api provides HTTP handlers for the pack allocation service.
// It implements a RESTful API for calculating pack distributions
// and retrieving allocation history.
package api

import (
	"net/http"
	"strconv"

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
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid quantity: " + err.Error(),
		})
		return
	}

	packs, _, err := h.allocator.CalculatePacks(quantity)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"packs": packs,
	})
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
