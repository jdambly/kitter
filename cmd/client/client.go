package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/jdambly/kitter/pkg/netapi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// NewCmd
func NewCmd() *cobra.Command {

	var server string
	var port string
	var wait time.Duration

	// create the "client" command
	cmd := &cobra.Command{
		Use:   "client",
		Short: "start client",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if the "server" flag was set
			if server == "" {
				return errors.New("the --server flag is required")
			}

			cNames, err := ResolveHostname(server)
			if err != nil {
				log.Error().Err(err).Msg("could not resolve hostname")
				return err
			}
			log.Debug().Strs("cnames", cNames)

			ticker := time.NewTimer(0)
			for {
				select {
				case <-cmd.Context().Done():
					log.Debug().Msg("Received termination signal")
					ticker.Stop()
					return nil
				case <-ticker.C:
					ConnectToMultipleServers(cNames, port)
					ticker.Reset(wait)
				}
			}

			return nil
		},
	}

	// Add the "host" flag to the "client" command.
	cmd.Flags().StringVarP(&server, "server", "s", "", "Host to connect to")
	cmd.Flags().StringVarP(&port, "port", "p", "5102", "Port to connect to")
	cmd.Flags().DurationVarP(&wait, "wait", "w", 1*time.Second, "Time in seconds to wait between polls")

	// Return the new command.
	return cmd
}

// connectToServer
func connectToServer(addr string, data string, ch chan string) error {
	log.Debug().Str("addr", addr).Msg("connecting to host")
	client := netapi.NewClient(addr)
	err := client.Connect()
	if err != nil {
		ch <- fmt.Sprintf("Failed to connect to %s: %v", addr, err)
		return err
	}
	defer client.Close()

	response, err := client.SendData(data)
	if err != nil {
		ch <- fmt.Sprintf("Failed to send data to %s: %v", addr, err)
		return err
	}

	ch <- fmt.Sprint(response)
	return nil
}

// ConnectToMultipleServers
func ConnectToMultipleServers(addresses []string, port string) {
	ch := make(chan string, len(addresses)) // Buffered channel to collect responses

	for _, addr := range addresses {
		go connectToServer(addr+":"+port, time.Now().Format(time.RFC3339Nano), ch) // Start a goroutine for each server connection
	}

	// Collect responses from all goroutines
	for i := 0; i < len(addresses); i++ {
		err := ProcessResponse(<-ch)
		if err != nil {
			log.Error().Err(err).Msg("")
		}
	}
}

// ResolveHostname resolves a given hostname to its IP addresses.
func ResolveHostname(hostname string) ([]string, error) {
	ips, err := net.LookupHost(hostname)
	if err != nil {
		// Check for specific DNS errors
		if dnsErr, ok := err.(*net.DNSError); ok {
			if dnsErr.Timeout() {
				return nil, errors.New("DNS query timed out")
			}
			if dnsErr.Temporary() {
				return nil, errors.New("Temporary DNS failure")
			}
			if strings.Contains(dnsErr.Error(), "no such host") {
				return nil, errors.New("Service not found in DNS")
			}
		}
		return nil, err
	}
	return ips, nil
}

// ProcessResponse take the response from the server and calculates the RRT latency
func ProcessResponse(data string) error {
	dStamp := time.Now()
	var resp netapi.Response
	err := json.Unmarshal([]byte(data), &resp)
	if err != nil {
		return err
	}
	resp.ClientDone = dStamp.Format(time.RFC3339Nano)
	cStamp, err := time.Parse(time.RFC3339Nano, resp.ClientTime)
	if err != nil {
		return err
	}

	rtt := dStamp.Sub(cStamp)
	resp.RTT = rtt.Milliseconds()
	log.Info().Any("response", resp).Msg("")
	return nil
}
