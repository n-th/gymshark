package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"

	"github.com/gin-gonic/gin"
	_ "github.com/n-th/gymshark/docs" // generated swagger docs
	"github.com/n-th/gymshark/internal/allocator"
	"github.com/n-th/gymshark/internal/api"
	"github.com/n-th/gymshark/internal/storage"
)

type Config struct {
	PackSizes []int `yaml:"pack_sizes"`
	Server    struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`
}

func loadConfig(path string) (*Config, error) {
	log.Printf("Loading config from %s", path)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var cfg Config
	dec := yaml.NewDecoder(f)
	if err := dec.Decode(&cfg); err != nil {
		return nil, err
	}
	log.Printf("Loaded config: pack_sizes=%v, server.host=%s, server.port=%d", cfg.PackSizes, cfg.Server.Host, cfg.Server.Port)
	return &cfg, nil
}

// @title Smart Pack Allocation API
// @version 1.0
// @description A Go-based API service that calculates optimal pack distribution for fulfilling orders with fixed pack sizes.
// @host localhost:8080
// @BasePath /
func main() {
	// Create data directory if it doesn't exist
	dataDir := "data"
	if os.Getenv("APP_ENV") == "docker" {
		dataDir = "/app/data"
	}
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Initialize storage
	store, err := storage.NewSQLiteStorage(filepath.Join(dataDir, "allocations.db"))
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize allocator with storage
	alloc := allocator.NewAllocator([]int{23, 31, 53}, store)
	defer alloc.Close()

	// Create a new Gin router
	router := gin.Default()

	// Create a new handler
	handler := api.NewHandler(alloc)

	// Register the routes
	handler.RegisterRoutes(router)

	// Create a new HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// Start the server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
