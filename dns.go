package main

import (
	"github.com/miekg/dns"
	"net"
	"strings"
)

// Lookup the ip address using HypeDNS
func lookup(hostname string) (response string, err error) {
	c := new(dns.Client)

	m := new(dns.Msg)
	m.SetQuestion(dns.Fqdn(hostname), dns.TypeAAAA)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, "[fc5d:baa5:61fc:6ffd:9554:67f0:e290:7535]:53")
	if r == nil || err != nil {
		return
	}

	// Stuff must be in the answer section
	for _, a := range r.Answer {
		columns := strings.Fields(a.String()) //column 4 holds the ip address
		return padIPv6(net.ParseIP(columns[4])), nil
	}
	return
}
