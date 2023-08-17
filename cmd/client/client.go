package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jasonhancock/go-http"
	"github.com/jdambly/kitter/pkg/netapi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var (
	metricRTT = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "kitter_rtt",
		Help:    "round trip time",
		Buckets: prometheus.DefBuckets, // default buckets
	}, []string{"target"})
)

const (
	maxRetries      = 5               // maximum number of retries
	initialWaitTime = 2 * time.Second // initial wait time before retry
	factor          = 5               // factor by which wait time increases
)

// NewCmd
func NewCmd() *cobra.Command {
	// create the vars I need to make sure they have the correct scope and are not shadowed
	var server string
	var port string
	var wait time.Duration
	var httpAddr string

	// create the "client" command
	cmd := &cobra.Command{
		Use:   "client",
		Short: "start client",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if the "server" flag was set
			if server == "" {
				return errors.New("the --server flag is required")
			}
			var cNames []string
			var err error
			for i := 0; i < maxRetries; i++ {
				cNames, err = ResolveHostname(server)
				if err == nil {
					break // if successful, break out of the loop
				}

				if i < maxRetries-1 { // don't sleep after the last attempt
					waitTime := time.Duration(int64(initialWaitTime) * int64(factor^i))
					log.Warn().Err(err).Dur("waitTime", waitTime).
						Msg("failed to resolve hostname, retrying...")
					time.Sleep(waitTime)
				} else {
					log.Error().Err(err).Msg("could not resolve hostname after multiple attempts")
					return err
				}
			}
			log.Debug().Strs("cnames", cNames).Msg("")

			registry, ok := prometheus.DefaultRegisterer.(*prometheus.Registry)
			if !ok {
				return errors.New("prometheus default registry is not a *prometheus.Registry")
			}

			var wg sync.WaitGroup
			router := chi.NewRouter()
			router.Mount("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
			http.NewHTTPServer(cmd.Context(), nil, &wg, router, httpAddr)

			ticker := time.NewTimer(0)
			dTicker := time.NewTicker(30)
			for {
				select {
				case <-cmd.Context().Done():
					log.Debug().Msg("received termination signal")
					ticker.Stop()
					return nil
				case <-ticker.C:
					ConnectToMultipleServers(cNames, port)
					ticker.Reset(wait)
				case <-dTicker.C:
					newNames, err := ResolveHostname(server)
					if err == nil {
						cNames = newNames
					}
				}
			}
		},
	}

	// Add the "host" flag to the "client" command.
	cmd.Flags().StringVarP(&server, "server", "s", "", "Host to connect to")
	cmd.Flags().StringVarP(&port, "port", "p", "5102", "Port to connect to")
	cmd.Flags().DurationVarP(&wait, "wait", "w", 1*time.Second, "Time in seconds to wait between polls")
	cmd.Flags().StringVar(&httpAddr, "http-addr", ":8080", "interface:port to bind the http server to")

	// Return the new command.
	return cmd
}

// connectToServer
func connectToServer(addr string, data string) (string, error) {
	log.Debug().Str("addr", addr).Msg("connection to hose")
	client := netapi.NewClient(addr)
	err := client.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer client.Close()

	response, err := client.SendData(data)
	if err != nil {
		return "", fmt.Errorf("failed to send data to %s: %w", addr, err)
	}

	return response, nil
}

// ConnectToMultipleServers
func ConnectToMultipleServers(addresses []string, port string) {
	ch := make(chan string, len(addresses)) // Buffered channel to collect responses

	// Collect responses from all goroutines
	go func() {
		for msg := range ch { // range over the channels (this is a good pattern)
			err := ProcessResponse(msg)
			if err != nil {
				log.Error().Err(err).Msg("processing response")
			}
		}
	}()

	var wg sync.WaitGroup
	for _, addr := range addresses {
		// Start a goroutine for each server connection
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			resp, err := connectToServer(addr+":"+port, time.Now().Format(time.RFC3339Nano))
			if err != nil {
				log.Error().Str("addr", addr+":"+port).Err(err).Msg("")
				return
			}
			ch <- resp
		}(addr)
	}
	wg.Wait() // wait for all the channels to do a thing and finish
	close(ch)
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
	resp.RTT = rtt.Seconds()
	metricRTT.WithLabelValues(resp.Server).Observe(resp.RTT)
	log.Info().Any("resp", resp).Msg("")

	return nil
}
