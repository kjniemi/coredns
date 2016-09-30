package cache

import (
	"testing"
	"time"

	"github.com/mholt/caddy"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		input        string
		shouldErr    bool
		expectedNcap int
		expectedPcap int
		execptedNttl time.Duration
		execptedPttl time.Duration
	}{
		{`cache`, false, defaultCap, defaultCap, defaultTTL, defaultTTl},
		{`cache {}`, false, defaultCap, defaultCap, defaultTTL, defaultTTl},
		{`cache aaa example.nl`, true, defaultCap, defaultCap, defaultTTL, defaultTTl},
	}
	for i, test := range tests {
		c := caddy.NewTestController("dns", test.input)
		ca, err := cacheParse(c)
		if test.shouldErr && err == nil {
			t.Errorf("Test %v: Expected error but found nil", i)
			continue
		} else if !test.shouldErr && err != nil {
			t.Errorf("Test %v: Expected no error but found error: %v", i, err)
			continue
		}

		if ca.ncap != test.expectedNcap {
			t.Errorf("Test %v: Expected ncap %v but found: %v", i, test.expectedNcap, ca.ncap)
		}
		if ca.pcap != test.expectedPcap {
			t.Errorf("Test %v: Expected pcap %v but found: %v", i, test.expectedPcap, ca.pcap)
		}
		if ca.nttl != test.expectedNttl {
			t.Errorf("Test %v: Expected nttl %v but found: %v", i, test.expectedNttl, ca.nttl)
		}
		if ca.pttl != test.expectedPttl {
			t.Errorf("Test %v: Expected pttl %v but found: %v", i, test.expectedPttl, ca.pttl)
		}
	}
}
