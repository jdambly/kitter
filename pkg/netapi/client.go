package netapi

import (
	"bufio"
	"errors"
	"net"
	"strings"
	"time"
)

// TCPClient is a struct that represents a TCP Client.
// It contains the address of the server and a connection to the server.
type TCPClient struct {
	// addr is the address of the server.
	addr string

	// conn is a netapi.Conn which represents the Client's connection to the server.
	conn net.Conn
}

// NewClient is a factory function that creates a new Client with the provided address. Example address localhost:8080
func NewClient(addr string) *TCPClient {
	// Create and return a new TCPClient with the provided address
	return &TCPClient{
		addr: addr,
	}
}

// Connect is a method on the TCPClient struct that establishes a connection to the server.
// It returns an error if any issues occur during the process.
func (c *TCPClient) Connect() (err error) {
	// Dial the server
	c.conn, err = net.Dial("tcp", c.addr)
	// If there is an error in dialing, return the error
	if err != nil {
		return
	}
	// Return nil if the connection was successful
	return nil
}

// SendData is a method on the TCPClient struct that sends data to the server.
// It returns the server's response and an error if any issues occur during the process.
func (c *TCPClient) SendData(data string) (string, error) {
	// Check if the connection is established
	if c.conn == nil {
		return "", errors.New("connection not established")
	}
	// Create a new reader and writer for the connection
	reader := bufio.NewReader(c.conn)
	writer := bufio.NewWriter(c.conn)

	// set the timeout on the connection
	err := c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	if err != nil {
		return "", err
	}

	// Write the data to the connection
	_, err = writer.WriteString(data + "\n")
	if err != nil {
		return "", err
	}

	err = writer.Flush()
	// If there is an error in writing, return the error
	if err != nil {
		return "", err
	}

	// Read the response from the connection
	response, err := reader.ReadBytes('\n')
	// If there is an error in reading, return the error
	if err != nil {
		return "", err
	}

	// Return the response and nil if the operation was successful
	return strings.Trim(string(response), "\n"), nil
}

// Close is a method on the TCPClient struct that shuts down the connection to the server.
// It returns an error if any issues occur during the process.
func (c *TCPClient) Close() error {
	// Check if the connection is established
	if c.conn == nil {
		return errors.New("connection not established")
	}

	// Close the connection
	err := c.conn.Close()
	// If there is an error in closing, return the error
	if err != nil {
		return err
	}

	// Return nil if the operation was successful
	return nil
}
