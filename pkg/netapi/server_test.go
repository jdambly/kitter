package netapi

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_ProcessData(t *testing.T) {
	// todo I don't really need to start the tcp server here I should be able to use the interface to mock it and just test processData
	// Setup the server
	srv, err := NewServer("tcp", ":1123")
	require.NoError(t, err, "error starting TCP server")
	// create a chanel used to send the ready signal
	readCh := make(chan struct{})

	// Run the server in Goroutine to stop tests from blocking test execution.
	go func() {
		err := srv.Run(readCh)
		assert.NoError(t, err)
	}()
	<-readCh // make sure the server is really started so that we don't have race conditions
	defer func(srv Server) {
		err := srv.Close()
		assert.NoError(t, err)
	}(srv) // Ensure the server is closed after the test

	// Create an instance of JitterServer
	tcp := &TCPServer{
		Addr:   "localhost",
		Client: "0.0.0.0",
	}

	// Generate a test timestamp, subtracting 10 milliseconds to simulate a delay
	testTimestamp := time.Now().Add(-10 * time.Millisecond).Format(time.RFC3339Nano)

	// Call ProcessData
	respBytes, err := tcp.ProcessData([]byte(testTimestamp))
	require.NoError(t, err, "ProcessData failed")

	// Unmarshal the response
	var resp Response
	require.NoError(t, json.Unmarshal(respBytes, &resp), "Failed to unmarshal response")

	// Validate the response
	assert.NotEqual(t, resp.ClientTime, resp.ServerTime)
	assert.GreaterOrEqual(t, float64(10), resp.Latency)
}
