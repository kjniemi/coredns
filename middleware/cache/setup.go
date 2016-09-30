package cache

import (
	"strconv"
	"time"

	"github.com/miekg/coredns/core/dnsserver"
	"github.com/miekg/coredns/middleware"

	"github.com/hashicorp/golang-lru"
	"github.com/mholt/caddy"
)

func init() {
	caddy.RegisterPlugin("cache", caddy.Plugin{
		ServerType: "dns",
		Action:     setup,
	})
}

// Cache sets up the root file path of the server.
func setup(c *caddy.Controller) error {
	ca, err := cacheParse(c)
	if err != nil {
		return middleware.Error("cache", err)
	}
	dnsserver.GetConfig(c).AddMiddleware(func(next middleware.Handler) middleware.Handler {
		ca.Next = next
		return ca
	})

	return nil
}

func cacheParse(c *caddy.Controller) (*Cache, error) {

	ca := &Cache{pcap: defaultCap, ncap: defaultCap, pttl: defaultTTL, nttl: defaultTTL}

	for c.Next() {
		// cache [ttl] [zones..]
		origins := make([]string, len(c.ServerBlockKeys))
		copy(origins, c.ServerBlockKeys)
		args := c.RemainingArgs()

		if len(args) > 0 {
			origins = args
			// first args may be just a number, then it is the ttl, if not it is a zone
			t := origins[0]
			ttl, err := strconv.Atoi(t)
			if len(args) > 1 && err != nil {
				// first arg should be number, but isn't
				return nil, err
			}
			if err == nil {
				origins = origins[1:]
				if len(origins) == 0 {
					// There was *only* the ttl, revert back to server block
					copy(origins, c.ServerBlockKeys)
				}
			}
		}
		// Refinements? In an extra block.
		for c.NextBlock() {
			switch c.Val() {
			// first number is cap, second is an new ttl
			case "positive":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return nil, c.ArgErr()
				}
				pcap, err := strconv.Atoi(args[0])
				if err != nil {
					return nil, err
				}
				ca.pcap = pcap
				if len(args) > 1 {
					pttl, err := strconv.Atoi(args[0])
					if err != nil {
						return nil, err
					}
					ca.pttl = time.Duration(pttl) * time.Second
				}
			case "negative":
				args := c.RemainingArgs()
				if len(args) == 0 {
					return nil, c.ArgErr()
				}
				ncap, err := strconv.Atoi(args[0])
				if err != nil {
					return nil, err
				}
				ca.ncap = ncap
				if len(args) > 1 {
					nttl, err := strconv.Atoi(args[0])
					if err != nil {
						return nil, err
					}
					ca.nttl = time.Duration(nttl) * time.Second
				}
			default:
				return nil, c.ArgErr()
			}
		}

		for i := range origins {
			origins[i] = middleware.Host(origins[i]).Normalize()
		}

		var err error
		ca.Zones = origins
		ca.pcache, err = lru.New(ca.pcap)
		if err != nil {
			return nil, err
		}
		ca.ncache, err = lru.New(ca.ncap)
		if err != nil {
			return nil, err
		}

		return ca, nil
	}

	return nil, nil
}
