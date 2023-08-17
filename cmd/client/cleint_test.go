package client

import (
	"encoding/json"
	"errors"
	"github.com/jdambly/kitter/pkg/netapi"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var counter int // use this to count the number of DNS responses
type MockDNSResolver struct{}

type myTest []struct {
	host     string
	expected []string
	err      bool
}

func (m *MockDNSResolver) LookupHost(host string) ([]string, error) {
	if host == "valid.com" {
		return []string{"192.168.1.1"}, nil
	}
	if host == "wait.com" {
		counter++
		if counter > 4 {
			return []string{"192.168.1.1"}, nil
		}
	}
	return nil, errors.New("invalid host")
}

func TestResolveHostname(t *testing.T) {
	mockResolver := &MockDNSResolver{}

	tests := myTest{
		{
			host:     "valid.com",
			expected: []string{"192.168.1.1"},
			err:      false,
		},
		{
			host:     "invalid.com",
			expected: nil,
			err:      true,
		},
	}

	for _, tt := range tests {
		result, err := ResolveHostname(mockResolver, tt.host)
		if tt.err {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}
	}
}

func TestWaitForDNS(t *testing.T) {
	mockResolver := &MockDNSResolver{}

	tests := myTest{
		{
			host:     "valid.com",
			expected: []string{"192.168.1.1"},
			err:      false,
		},
		{
			host:     "invalid.com",
			expected: nil,
			err:      true,
		},
		{
			host:     "wait.com",
			expected: []string{"192.168.1.1"},
			err:      false,
		},
	}

	retries := RetryConfig{
		maxRetries:      3,
		initialWaitTime: 2 * time.Nanosecond,
		factor:          5,
	}

	for _, tt := range tests {
		result, err := WaitForDNS(mockResolver, retries, tt.host)
		if tt.err {
			assert.Error(t, err)
		}
		if err != nil && result != nil {
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		}

	}
}

func TestProcessResponse(t *testing.T) {
	resp := netapi.Response{
		ServerTime: time.Now().Add(-5 * time.Millisecond).Format(time.RFC3339Nano),
		ClientTime: time.Now().Format(time.RFC3339Nano),
	}
	data, _ := json.Marshal(resp)

	err := ProcessResponse(string(data))
	assert.NoError(t, err)
}
