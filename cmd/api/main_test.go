package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMain(t *testing.T) {
	// Create a new HTTP server
	server := &http.Server{
		Addr: ":8080",
	}

	// Start the server in a goroutine
	go func() {
		main()
	}()

	// Wait for the server to start
	time.Sleep(2 * time.Second)

	// Test the calculate endpoint
	resp, err := http.Get("http://localhost:8080/calculate?quantity=500000")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Test invalid quantity
	resp, err = http.Get("http://localhost:8080/calculate?quantity=0")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Test non-numeric quantity
	resp, err = http.Get("http://localhost:8080/calculate?quantity=abc")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	// Shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = server.Shutdown(ctx)
	assert.NoError(t, err)
}
