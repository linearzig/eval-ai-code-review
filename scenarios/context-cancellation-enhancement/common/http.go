package common

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// HTTPHandler interface for different types of handlers
type HTTPHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// BasicHandler implements basic HTTP handling
type BasicHandler struct {
	name string
}

func (h *BasicHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// Process HTTP request with standard handling
	
	// Process request without context awareness
	processRequestWithoutContext(w, r)
}

// LongRunningHandler implements long-running HTTP operations
type LongRunningHandler struct {
	name string
}

func (h *LongRunningHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// Handle long-running HTTP operations
	
	// Start long-running operation without context
	go func() {
		// This goroutine will run for the duration of the operation
		time.Sleep(30 * time.Second) // Simulate long operation
		
		// Try to write response (may fail if client disconnected)
		w.Write([]byte("Long operation completed"))
	}()
	
	// Return immediately without waiting for context cancellation
}

// DatabaseHandler implements database operations
type DatabaseHandler struct {
	name string
}

func (h *DatabaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This demonstrates missing context propagation in database operations
	
	// Perform database operation
	performDatabaseOperationWithoutContext(w, r)
}

// FileHandler implements file operations
type FileHandler struct {
	name string
}

func (h *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This demonstrates missing context propagation in file operations
	
	// Perform file operation
	performFileOperationWithoutContext(w, r)
}

// NetworkHandler implements network operations
type NetworkHandler struct {
	name string
}

func (h *NetworkHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This demonstrates missing context propagation in network operations
	
	// Perform network operation
	performNetworkOperationWithoutContext(w, r)
}

// processRequestWithoutContext processes a request without context awareness
func processRequestWithoutContext(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context usage
	
	// Simulate some processing time
	time.Sleep(5 * time.Second)
	
	// Write response
	w.Write([]byte("Request processed"))
}

// performDatabaseOperationWithoutContext performs database operations without context
func performDatabaseOperationWithoutContext(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in database operations
	
	// Simulate database query
	time.Sleep(10 * time.Second)
	
	// Write response
	w.Write([]byte("Database operation completed"))
}

// performFileOperationWithoutContext performs file operations without context
func performFileOperationWithoutContext(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in file operations
	
	// Simulate file operation
	time.Sleep(15 * time.Second)
	
	// Write response
	w.Write([]byte("File operation completed"))
}

// performNetworkOperationWithoutContext performs network operations without context
func performNetworkOperationWithoutContext(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in network operations
	
	// Simulate network request
	time.Sleep(20 * time.Second)
	
	// Write response without checking if client is still connected
	w.Write([]byte("Network operation completed"))
}

// StartHTTPServer starts an HTTP server with various handlers
func StartHTTPServer(addr string) error {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context propagation in server setup
	
	mux := http.NewServeMux()
	
	// Register handlers that don't use context properly
	mux.Handle("/basic", &BasicHandler{name: "basic"})
	mux.Handle("/long", &LongRunningHandler{name: "long"})
	mux.Handle("/db", &DatabaseHandler{name: "database"})
	mux.Handle("/file", &FileHandler{name: "file"})
	mux.Handle("/network", &NetworkHandler{name: "network"})
	
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	
	// Start server
	return server.ListenAndServe()
}

// ProcessRequestWithTimeout processes a request with a timeout but without context
func ProcessRequestWithTimeout(w http.ResponseWriter, r *http.Request, timeout time.Duration) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates using timeout without context
	

	
	// Create a timer for timeout
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	
	// Process request
	done := make(chan bool, 1)
	go func() {
		// Simulate processing
		time.Sleep(5 * time.Second)
		done <- true
	}()
	
	// Wait for either completion or timeout
	select {
	case <-done:
		w.Write([]byte("Request completed"))
	case <-timer.C:
		w.WriteHeader(http.StatusRequestTimeout)
		w.Write([]byte("Request timeout"))
	}
}

// HandleWebSocket handles WebSocket connections without context
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in WebSocket handling
	

	
	// Simulate WebSocket upgrade
	time.Sleep(2 * time.Second)
	
	// Start WebSocket handling without context
	go func() {
		// This goroutine will run indefinitely
		// even if the client disconnects
		for {
			time.Sleep(time.Second)
			// Process WebSocket messages
		}
	}()
	
	w.Write([]byte("WebSocket connection established"))
}

// StreamData streams data without context awareness
func StreamData(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in streaming operations
	

	
	// Set headers for streaming
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Transfer-Encoding", "chunked")
	
	// Stream data without checking context
	for i := 0; i < 100; i++ {
		// Simulate data generation
		time.Sleep(100 * time.Millisecond)
		
		// Write chunk without checking if client is still connected
		fmt.Fprintf(w, "Chunk %d\n", i)
		
		// Flush the response writer
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

// ProcessBatchRequest processes batch requests without context
func ProcessBatchRequest(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in batch processing
	

	
	// Simulate batch processing
	items := []string{"item1", "item2", "item3", "item4", "item5"}
	
	for _, item := range items {
		// Process each item without checking context
		time.Sleep(2 * time.Second)
		
		// Write progress without checking if client is still connected
		fmt.Fprintf(w, "Processed: %s\n", item)
		
		// Flush the response writer
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
	
	w.Write([]byte("Batch processing completed"))
}

// HandleUpload handles file uploads without context
func HandleUpload(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in file uploads
	

	
	// Parse multipart form without context
	err := r.ParseMultipartForm(32 << 20) // 32 MB
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	
	// Process uploaded files without context
	files := r.MultipartForm.File["files"]
	for _, fileHeader := range files {
		// Simulate file processing
		time.Sleep(3 * time.Second)
		
		// Write progress without checking if client is still connected
		fmt.Fprintf(w, "Processed file: %s\n", fileHeader.Filename)
	}
	
	w.Write([]byte("Upload completed"))
}

// HandleDownload handles file downloads without context
func HandleDownload(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in file downloads
	

	
	// Set headers for file download
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename=large-file.dat")
	
	// Simulate large file download
	for i := 0; i < 1000; i++ {
		// Generate chunk of data
		chunk := make([]byte, 1024)
		for j := range chunk {
			chunk[j] = byte(i % 256)
		}
		
		// Write chunk without checking if client is still connected
		w.Write(chunk)
		
		// Simulate processing time
		time.Sleep(10 * time.Millisecond)
		
		// Flush the response writer
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
	}
}

// HandleAPIRequest handles API requests without context
func HandleAPIRequest(w http.ResponseWriter, r *http.Request) {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in API requests
	

	
	// Simulate API processing
	time.Sleep(5 * time.Second)
	
	// Make external API call without context
	externalAPICallWithoutContext()
	
	// Write response without checking if client is still connected
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "API request completed"}`))
}

// externalAPICallWithoutContext makes an external API call without context
func externalAPICallWithoutContext() {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in external API calls
	

	
	// Simulate external API call
	time.Sleep(3 * time.Second)
	
	// Process API response without context
	processAPIResponseWithoutContext()
}

// processAPIResponseWithoutContext processes API response without context
func processAPIResponseWithoutContext() {
	// Enhancement: Add request cancellation support for HTTP endpoints
	// This function demonstrates missing context in response processing
	

	
	// Simulate response processing
	time.Sleep(2 * time.Second)
} 