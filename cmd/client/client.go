package client

import (
	"errors"
	"fmt"
	"github.com/jdambly/kitter/pkg/netapi"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"net"
	"time"
)

// NewCmd
func NewCmd() *cobra.Command {
	// create the "client" command
	cmd := &cobra.Command{
		Use:   "client",
		Short: "start client",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if the "server" flag was set
			if !cmd.Flag("server").Changed {
				return errors.New("the --server flag is required")
			}

			// get the value of the "host" flag
			host, _ := cmd.Flags().GetString("server")
			port, _ := cmd.Flags().GetString("port")
			cNames, err := ResolveHostname(host)
			if err != nil {
				log.Error().Err(err).Msg("could not resolve hostname")
				return err
			}
			log.Debug().Strs("cnames", cNames)
			ConnectToMultipleServers(cNames, port)
			return nil
		},
	}
	// Add the "host" flag to the "client" command.
	cmd.Flags().StringP("server", "s", "", "Host to connect to")
	cmd.Flags().StringP("port", "p", "5102", "Port to connect to")

	// Return the new command.
	return cmd
}

// connectToServer
func connectToServer(addr string, data string, ch chan string) {
	log.Debug().Str("addr", addr).Msg("connecting to host")
	client := netapi.NewClient(addr)
	err := client.Connect()
	if err != nil {
		ch <- fmt.Sprintf("Failed to connect to %s: %v", addr, err)
		return
	}
	defer client.Close()

	response, err := client.SendData(data)
	if err != nil {
		ch <- fmt.Sprintf("Failed to send data to %s: %v", addr, err)
		return
	}

	ch <- fmt.Sprint(response)
}

// ConnectToMultipleServers
func ConnectToMultipleServers(addresses []string, port string) {
	ch := make(chan string, len(addresses)) // Buffered channel to collect responses

	for _, addr := range addresses {
		go connectToServer(addr+":"+port, time.Now().Format(time.RFC3339Nano), ch) // Start a goroutine for each server connection
	}

	// Collect responses from all goroutines
	for i := 0; i < len(addresses); i++ {
		fmt.Println(<-ch)
	}
}

// ResolveHostname resolves a given hostname to its IP addresses.
func ResolveHostname(hostname string) ([]string, error) {
	// LookupIP looks up host using the local resolver.
	// It returns a slice of that host's IPv4 and/or IPv6 addresses.
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve hostname %s: %v", hostname, err)
	}

	var ipStrings []string
	for _, ip := range ips {
		ipStrings = append(ipStrings, ip.String())
	}

	return ipStrings, nil
}
