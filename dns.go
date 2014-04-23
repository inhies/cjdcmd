/*
 * You may redistribute this program and/or modify it under the terms of
 * the GNU General Public License as published by the Free Software Foundation,
 * either version 3 of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */
package main

import (
	"fmt"
	"github.com/miekg/dns"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
)

var NoDNS bool

// Lookup the IP address using HypeDNS
func lookupHypeDNS(hostname string) (response string, err error) {
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

// Lookup the hostname for an IP address using HypeDNS
// This probably needs work but I don't really know what I'm doing :)
func reverseHypeDNSLookup(ip string) (response string, err error) {
	c := new(dns.Client)

	m := new(dns.Msg)
	thing, err := dns.ReverseAddr(ip)
	if err != nil {
		return
	}
	m.SetQuestion(thing, dns.TypePTR)
	m.RecursionDesired = true

	r, _, err := c.Exchange(m, "[fc5d:baa5:61fc:6ffd:9554:67f0:e290:7535]:53")
	if r == nil || err != nil {
		return
	}

	// Stuff must be in the answer section
	for _, a := range r.Answer {
		columns := strings.Fields(a.String()) //column 4 holds the ip address
		return columns[4], nil
	}
	return
}

// Create a HypeDNS name for this device
func setHypeDNS(hostname string) (response string, err error) {
	setLoc := "/_hypehost/set?hostname="
	getLoc := "/_hypehost/get"
	nodeInfoHost := "http://[fc5d:baa5:61fc:6ffd:9554:67f0:e290:7535]:8000"

	if len(hostname) == 0 {
		// The request is a "get"
		resp, err := http.Get(nodeInfoHost + getLoc)
		if err != nil {
			fmt.Println("Got an error, %s", err)
			err = fmt.Errorf("Got an error when attempting to retrieve " +
				"hostname. This is usually because you can't connect to HypeDNS. " +
				"Try again later")
			return "", err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		fmt.Println("You are: " + string(body))
		return "", nil
	}
	// The request is a "set"
	resp, err := http.Get(nodeInfoHost + setLoc + hostname)
	if err != nil {
		fmt.Println("Got an error, %s", err)
		err = fmt.Errorf("Got an error when attempting to change hostname. " +
			"Try again later")
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	fmt.Println("Hostname " + string(body) + " created.")
	return "", nil
}

// Resolve an IP to a domain name using the system DNS settings first, then HypeDNS
func resolveIP(ip string) (hostname string, err error) {
	var try2 string
	if NoDNS {
		hostname = ip
	} else {
		// try the system DNS setup
		result, _ := net.LookupAddr(ip)
		if len(result) > 0 {
			goto end
		}

		// Try HypeDNS
		try2, err = reverseHypeDNSLookup(ip)
		if try2 == "" || err != nil {
			err = fmt.Errorf("Unable to resolve IP address. This is usually caused by not having a route to hypedns. Please try again in a few seconds.")
			return
		}
		result = append(result, try2)
	end:
		for _, addr := range result {
			hostname = addr
		}
	}
	// Trim the trailing period becuase it annoys me
	if hostname[len(hostname)-1] == '.' {
		hostname = hostname[:len(hostname)-1]
	}
	return
}

// Resolve a hostname to an IP address using the system DNS settings first, then HypeDNS
func resolveHost(hostname string) (ips []string, err error) {
	var ip string
	// Try the system DNS setup
	result, _ := net.LookupHost(hostname)
	if len(result) > 0 {
		goto end
	}

	// Try with hypedns
	ip, err = lookupHypeDNS(hostname)

	if ip == "" || err != nil {
		err = fmt.Errorf("Unable to resolve hostname. This is usually caused by not having a route to hypedns. Please try again in a few seconds.")
		return
	}

	result = append(result, ip)

end:
	for _, addr := range result {
		tIP := net.ParseIP(addr)
		// Only grab the cjdns IP's
		if tIP[0] == 0xfc {
			ips = append(ips, padIPv6(net.ParseIP(addr)))
		}
	}

	return
}
