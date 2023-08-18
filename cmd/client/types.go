package client

import (
	"net"
	"time"
)

type DNSResolver interface {
	LookupHost(host string) ([]string, error)
}

type DefaultDNSResolver struct{}

func (d *DefaultDNSResolver) LookupHost(host string) ([]string, error) {
	return net.LookupHost(host)
}

type RetryConfig struct {
	maxRetries      int           // maximum number of retries
	initialWaitTime time.Duration // initial wait time before retry
	factor          int           // factor by which wait time increases
}
