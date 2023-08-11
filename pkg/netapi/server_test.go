package netapi

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"log"
	"time"

	"testing"
)

var srv Server
var err error

func init() {
	// Start the new server.
	srv, err = NewServer("tcp", ":1123")
	if err != nil {
		log.Println("error starting TCP server")
		return
	}

	// Run the server in Goroutine to stop tests from blocking
	// test execution.
	go func() {
		_ = srv.Run()
	}()
}

func Test_ProcessData(t *testing.T) {
	// Create an instance of JitterServer
	tcp := &TCPServer{
		Addr:   "localhost",
		Client: "0.0.0.0",
	}

	// Generate a test timestamp, subtracting 10 milliseconds to simulate a delay
	testTimestamp := time.Now().Add(-10 * time.Millisecond).Format(time.RFC3339Nano)

	// Call ProcessData
	respBytes, err := tcp.ProcessData([]byte(testTimestamp))
	if err != nil {
		t.Fatalf("ProcessData failed: %v", err)
	}

	// Unmarshal the response
	var resp Response
	if err := json.Unmarshal(respBytes, &resp); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Validate the response
	assert.NotEqual(t, resp.ClientTime, resp.ServerTime)
	assert.Equal(t, int64(10), resp.Latency)
}
