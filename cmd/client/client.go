package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	clog "github.com/jasonhancock/cobra-logger"
	"github.com/jasonhancock/go-http"
	"github.com/jasonhancock/go-logger"
	"github.com/jdambly/kitter/pkg/netapi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	factor          = 2               // factor by which wait time increases
)

// NewCmd
func NewCmd() *cobra.Command {

	var server string
	var port string
	var wait time.Duration
	var httpAddr string
	var logConf *clog.Config

	// create the "client" command
	cmd := &cobra.Command{
		Use:   "client",
		Short: "start client",
		RunE: func(cmd *cobra.Command, args []string) error {
			l := logConf.Logger(os.Stdout)

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
					l.Warn("failed to resolve hostname, retrying...", "error", err, "waitTime", waitTime)
					time.Sleep(waitTime)
				} else {
					l.Err("could not resolve hostname after multiple attempts", "error", err)
					return err
				}
			}

			l.Debug("cnames", "cnames", strings.Join(cNames, ","))

			registry, ok := prometheus.DefaultRegisterer.(*prometheus.Registry)
			if !ok {
				return errors.New("prometheus default registry is not a *prometheus.Registry")
			}

			var wg sync.WaitGroup
			router := chi.NewRouter()
			router.Mount("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
			http.NewHTTPServer(cmd.Context(), l.New("http"), &wg, router, httpAddr)

			ticker := time.NewTimer(0)
			for {
				select {
				case <-cmd.Context().Done():
					l.Debug("received termination signal")
					ticker.Stop()
					return nil
				case <-ticker.C:
					ConnectToMultipleServers(l, cNames, port)
					ticker.Reset(wait)
				}
			}
		},
	}

	logConf = clog.NewConfig(cmd)

	// Add the "host" flag to the "client" command.
	cmd.Flags().StringVarP(&server, "server", "s", "", "Host to connect to")
	cmd.Flags().StringVarP(&port, "port", "p", "5102", "Port to connect to")
	cmd.Flags().DurationVarP(&wait, "wait", "w", 1*time.Second, "Time in seconds to wait between polls")
	cmd.Flags().StringVar(&httpAddr, "http-addr", ":8080", "interface:port to bind the http server to")

	// Return the new command.
	return cmd
}

// connectToServer
func connectToServer(l *logger.L, addr string, data string) (string, error) {
	l.Debug("connecting to host", "addr", addr)
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
func ConnectToMultipleServers(l *logger.L, addresses []string, port string) {
	ch := make(chan string, len(addresses)) // Buffered channel to collect responses

	// Collect responses from all goroutines
	go func() {
		for msg := range ch { // range over the channels (this is a good pattern)
			err := ProcessResponse(l, msg)
			if err != nil {
				l.Err("processing response", "error", err)
			}
		}
	}()

	var wg sync.WaitGroup
	for _, addr := range addresses {
		// Start a goroutine for each server connection
		wg.Add(1)
		go func(addr string) {
			defer wg.Done()
			resp, err := connectToServer(l, addr+":"+port, time.Now().Format(time.RFC3339Nano))
			if err != nil {
				l.Err("connecting to server", "addr", addr+":"+port, "error", err)
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
func ProcessResponse(l *logger.L, data string) error {
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
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	l.Info("response", "response", string(b))

	return nil
}
