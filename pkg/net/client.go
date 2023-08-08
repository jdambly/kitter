package net

import (
	"bufio"
	"errors"
	"net"
)

// Client is an interface that defines methods for connecting to a server, sending data, and closing the connection.
type Client interface {
	// Connect establishes a connection to the server and returns an error if any issues occur during the process.
	Connect() error

	// SendData sends data to the server and returns the server's response and an error if any issues occur during the process.
	SendData(data []byte) ([]byte, error)

	// Close shuts down the connection to the server and returns an error if any issues occur during the process.
	Close() error
}

// TCPClient is a struct that represents a TCP client.
// It contains the address of the server and a connection to the server.
type TCPClient struct {
	// addr is the address of the server.
	addr string

	// conn is a net.Conn which represents the client's connection to the server.
	conn net.Conn
}

// NewClient is a factory function that creates a new Client with the provided address.
func NewClient(addr string) Client {
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
func (c *TCPClient) SendData(data []byte) ([]byte, error) {
	// Check if the connection is established
	if c.conn == nil {
		return nil, errors.New("connection not established")
	}

	// Write the data to the connection
	_, err := c.conn.Write(data)
	// If there is an error in writing, return the error
	if err != nil {
		return nil, err
	}

	// Create a new reader for the connection
	reader := bufio.NewReader(c.conn)

	// Read the response from the connection
	response, err := reader.ReadBytes('\n')
	// If there is an error in reading, return the error
	if err != nil {
		return nil, err
	}

	// Return the response and nil if the operation was successful
	return response, nil
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
