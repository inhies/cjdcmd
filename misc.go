package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/inhies/go-cjdns/admin"
	"net"
	"strings"
)

// Fills out an IPv6 address to the full 32 bytes
func padIPv6(ip net.IP) string {
	raw := hex.EncodeToString(ip)
	parts := make([]string, len(raw)/4)
	for i := range parts {
		parts[i] = raw[i*4 : (i+1)*4]
	}
	return strings.Join(parts, ":")
}

// Dumps the entire routing table and structures it
func getTable(user *admin.Admin) (table []*Route) {
	page := 0
	var more int64
	table = make([]*Route, 0)
	for more = 1; more != 0; page++ {
		response, err := admin.NodeStore_dumpTable(user, page)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		// If an error field exists, and we have an error, return it
		if _, ok := response["error"]; ok {
			if response["error"] != "none" {
				err = fmt.Errorf(response["error"].(string))
				fmt.Printf("Error: %v\n", err)
				return
			}
		}
		//Thanks again to SashaCrofter for the table parsing
		rawTable := response["routingTable"].([]interface{})
		for i := range rawTable {
			item := rawTable[i].(map[string]interface{})
			rPath := item["path"].(string)
			sPath := strings.Replace(rPath, ".", "", -1)
			bPath, err := hex.DecodeString(sPath)
			if err != nil || len(bPath) != 8 {
				//If we get an error, or the
				//path is not 64 bits, discard.
				//This should also prevent
				//runtime errors.
				continue
			}
			path := binary.BigEndian.Uint64(bPath)
			table = append(table, &Route{
				IP:      item["ip"].(string),
				RawPath: path,
				Path:    rPath,
				RawLink: item["link"].(int64),
				Link:    float64(item["link"].(int64)) / magicalLinkConstant,
				Version: item["version"].(int64),
			})

		}

		if response["more"] != nil {
			more = response["more"].(int64)
		} else {
			break
		}
	}

	return
}

func getTarget(input []string, allowHost bool, allowIP bool, allowPath bool) (target string, err error) {
	if len(input) <= 0 {
		err = fmt.Errorf("Invalid target specified")
		return
	}

	//Check to see if we sent an IPv6 address
	if err == nil && target == "" && allowIP && strings.Count(input[0], ":") > 2 {
		tempTarget := padIPv6(net.ParseIP(input[0]))
		if len(tempTarget) != 39 {
			err = fmt.Errorf("Invalid IPv6 address")
		} else {
			target = tempTarget
		}

	}

	//Check to see if we were sent a cjdns path
	if err == nil && target == "" && allowPath && strings.Count(input[0], ".") == 3 && len(input[0]) == 19 {
		tempTarget := input[0]
		valid := true
		for _, c := range tempTarget {
			if !strings.ContainsRune("abcdeABCDEF0123456789.", c) {
				valid = false
				break
			}
		}
		if valid {
			if len(tempTarget) != 19 {
				err = fmt.Errorf("Invalid cjdns path")
			} else {
				target = tempTarget
			}
		}
	}

	if err == nil && target == "" && allowHost {
		var ip string
		ip, err = lookup(input[0])
		if err != nil {
			return
		} else if ip == "" {
			err = fmt.Errorf("Unable to resolve hostname")
		} else {
			target = ip
			println("Resolved to:", ip)
		}
	}
	return

}
func usage() {
	println("cjdcmd version ", Version)
	println("")
	println("cjdcmd expects your cjdroute.conf to be at /etc/cjdroute.conf however you may specify")
	println("where to look with the -f or -file flags. It is recommended to make a symlink to your config")
	println("and place it at /etc/cjdroute to save on typing")
	println("")
	println("Usage: cjdcmd command [arguments]")
	println("")
	println("The commands are:")
	println("")
	println("ping <ipv6 address, hostname, or routing path>     sends a cjdns ping to the specified node")
	println("route <ipv6 address, hostname, or routing path>    prints out all routes to an IP or the IP to a route")
	println("traceroute <ipv6 address or hostname> [-t timeout] performs a traceroute by pinging each known hop to the target on all known paths")
	println("log [-l level] [-file file] [-line line]           prints cjdns log to stdout")
	println("peers                                              displays a list of currently connected peers")
	println("dump                                               dumps the routing table to stdout")
	println("kill                                               tells cjdns to gracefully exit")
	println("")
	println("Please use `cjdcmd --help` for a list of flags.")
}
