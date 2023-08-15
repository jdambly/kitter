package netapi

import (
	"bufio"
	"encoding/json"
	"errors"
	"github.com/rs/zerolog/log"
	"net"
	"os"
	"strings"
	"time"
)

// Server is an interface that defines methods for running and closing a server.
type Server interface {
	// Run starts the server and returns an error if any issues occur during the startup process.
	Run() error

	// Close shuts down the server and returns an error if any issues occur during the shutdown process.
	Close() error

	// ProcessData processes the data received from a Client connection and returns the response data and an error.
	// The implementation of this method can be overridden in different projects to provide custom data processing.
	ProcessData(data []byte) ([]byte, error)
}

// TCPServer is a struct that represents a TCP server.
// It contains the address of the server and a Listener from the netapi package
// that listens for incoming connections on the address.
type TCPServer struct {
	// Addr is the address where the server is hosted.
	Addr   string
	Client string

	// server is a netapi.Listener which accepts incoming connections on the Addr.
	server net.Listener
}

type Response struct {
	ServerTime string  `json:"serverTime"`
	ClientTime string  `json:"clientTime"`
	Client     string  `json:"client"`
	Server     string  `json:"server"`
	Latency    float64 `json:"latency"`
	ClientDone string  `json:"clientDone"`
	RTT        float64 `json:"RTT"`
}

// NewServer is a factory function that creates a new Server based on the provided protocol and address.
// Currently, it only supports TCP. If an unsupported protocol is provided, it returns an error.
func NewServer(protocol, addr string) (Server, error) {
	// Convert the protocol to lower case to ensure case-insensitive comparison
	switch strings.ToLower(protocol) {
	case "tcp":
		// If the protocol is TCP, create and return a new TCPServer with the provided address
		return &TCPServer{
			Addr: addr,
		}, nil
	case "udp":
		//  todo If the protocol is UDP implement it
	}
	// If the protocol is neither TCP nor UDP, return an error
	return nil, errors.New("invalid protocol given")
}

// Run is a method on the TCPServer struct that starts the TCP server.
// It listens for incoming connections on the server's address and accepts them.
// If there is an error in listening or accepting connections, it returns the error.
func (t *TCPServer) Run() (err error) {
	// Listen on the TCP network at the server's address
	t.server, err = net.Listen("tcp", t.Addr)
	// If there is an error in listening, return the error
	if err != nil {
		return
	}
	// Handle connections
	err = t.handleConnections()
	return err
}

// Close shuts down the TCP Server
func (t *TCPServer) Close() (err error) {
	return t.server.Close()
}

// handleConnections is a method on the TCPServer struct that accepts incoming connections and handles them concurrently.
func (t *TCPServer) handleConnections() (err error) {
	for {
		// Accept a new connection
		conn, err := t.server.Accept()
		if err != nil || conn == nil {
			// If there is an error in accepting the connection or the connection is nil, create a new error and break the loop
			err = errors.New("could not accept connection")
			break
		}

		// Handle the connection concurrently
		go t.handleConnection(conn)
	}
	return
}

// handleConnection is a method on the TCPServer struct that handles a single connection.
// It reads data from the connection, processes it using the ProcessData method, and writes the response back to the connection.
func (t *TCPServer) handleConnection(conn net.Conn) {
	// Close the connection when the function returns
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)
	// get the Client address
	t.Client = conn.LocalAddr().String()

	// Create a new reader and writer for the connection
	reader := bufio.NewReader(conn)
	writer := bufio.NewWriter(conn)

	// Read data from the connection
	data, err := reader.ReadBytes('\n')
	if err != nil {
		// If there is an error in reading, write an error message back to the connection and return
		_, _ = writer.WriteString("failed to read input")
		_ = writer.Flush()
		return
	}

	// Process the data
	response, err := t.ProcessData(data)
	if err != nil {
		// If there is an error in processing the data, write an error message back to the connection and return
		_, _ = writer.WriteString("failed to process data")
		_ = writer.Flush()
		return
	}

	// Write the response back to the connection
	_, _ = writer.Write(response)
	_, _ = writer.WriteString("\n")
	_ = writer.Flush()
}

/*
// ProcessData is a method on the TCPServer struct that processes the data received from a Client connection.
// This implementation simply echoes back the received data.
func (t *TCPServer) ProcessData(data []byte) ([]byte, error) {
	// Echo back the received data
	return data, nil
}
*/

func (t *TCPServer) ProcessData(data []byte) ([]byte, error) {
	// server response time
	sStamp := time.Now()
	// data should be a time.RFC3339Nano string
	// this is the client timestamp
	str := strings.TrimSuffix(string(data), "\n") // remove newline from the data
	log.Info().Str("data", str).Msg("")
	cStamp, err := time.Parse(time.RFC3339Nano, str)
	if err != nil {
		message := "Could not parse timestamp from client"
		log.Error().Err(err).Msg(message)
		return []byte(message), err
	}
	// Calculate the one way latency
	latency := sStamp.Sub(cStamp)
	// create the response
	resp := Response{
		ServerTime: sStamp.Format(time.RFC3339Nano),
		ClientTime: cStamp.Format(time.RFC3339Nano),
		Client:     t.Client,
		Latency:    latency.Seconds(),
	}
	getNamespaceFromEnv(&resp)
	// convert to json
	respBytes, err := json.Marshal(resp)
	if err != nil {
		log.Error().Err(err).Msg("could not marshal response")
		return []byte("could not marshal response"), nil
	}

	return respBytes, nil
}

func getNamespaceFromEnv(data *Response) {
	data.Server = os.Getenv("POD_NAME")
}
