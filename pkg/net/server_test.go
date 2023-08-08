package net

import (
	"bytes"
	"log"

	"net"
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

// TestNETServer_Request is a test function that tests the request handling of the TCP server.
// It sends a payload to the server and checks the response.
func TestNETServer_Request(t *testing.T) {
	// Define the test cases
	tt := []struct {
		test    string
		payload []byte
		want    []byte
	}{
		{
			"Sending a simple request returns result",
			[]byte("hello world\n"),
			[]byte("Request received: hello world"),
		},
		{
			"Sending another simple request works",
			[]byte("goodbye world\n"),
			[]byte("Request received: goodbye world"),
		},
	}

	// Iterate over the test cases
	for _, tc := range tt {
		// Run each test case as a subtest
		t.Run(tc.test, func(t *testing.T) {
			// Dial the TCP server
			conn, err := net.Dial("tcp", ":1123")
			if err != nil {
				t.Error("could not connect to TCP server: ", err)
			}
			defer func(conn net.Conn) {
				_ = conn.Close()
			}(conn)

			// Write the payload to the server
			if _, err := conn.Write(tc.payload); err != nil {
				t.Error("could not write payload to TCP server:", err)
			}

			// Read the response from the server
			out := make([]byte, 1024)
			if _, err := conn.Read(out); err == nil {
				// Compare the response with the expected output
				if bytes.Compare(out, tc.want) == 0 {
					t.Error("response did match expected output")
				}
			} else {
				t.Error("could not read from connection")
			}
		})
	}
}
