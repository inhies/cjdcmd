package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/inhies/go-cjdns/admin"
	"github.com/inhies/go-cjdns/config"
	"math/rand"
	"net"
	"regexp"
	"strings"
)

const (
	ipRegex   = "^fc[a-f0-9]{1,2}:([a-f0-9]{0,4}:){2,6}[a-f0-9]{1,4}$"
	pathRegex = "([0-9a-f]{4}\\.){3}[0-9a-f]{4}"
	hostRegex = "^([a-zA-Z0-9]([a-zA-Z0-9\\-\\.]{0,}[a-zA-Z0-9]))$"
)

func readConfig() (conf *config.Config, err error) {
	if AdminPassword == defaultPass {
		conf, err = config.LoadMinConfig(File)
		if err != nil || len(conf.Admin.Password) == 0 {
			return
		}

		AdminPassword = conf.Admin.Password
		AdminBind = conf.Admin.Bind
	} else {
		AdminBind = defaultAdminBind
	}
	return
}
func adminConnect() (user *admin.Admin, err error) {
	user, err = admin.Connect(AdminBind, AdminPassword)
	if err != nil {
		if e, ok := err.(net.Error); ok {
			if e.Timeout() {
				fmt.Println("\nConnection timed out")
			} else if e.Temporary() {
				fmt.Println("\nTemporary error (not sure what that means!)")
			} else {
				fmt.Println("\nUnable to connect to cjdns:", e)
			}
		} else {
			fmt.Println("\nError:", err)
		}
		return
	}
	return
}
func connect() (user *admin.Admin, err error) {
	_, err = readConfig()
	if err != nil {
		///fmt.Println(err)
		return
	}
	user, err = adminConnect()
	if err != nil {
		//fmt.Println(err)
		return
	}
	//defer user.Conn.Close()
	return
}

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

type Target struct {
	Target   string
	Supplied string
}

func validIP(input string) (result bool) {
	result, _ = regexp.MatchString(ipRegex, input)
	return
}
func validPath(input string) (result bool) {
	result, _ = regexp.MatchString(pathRegex, input)
	return
}
func validHost(input string) (result bool) {
	result, _ = regexp.MatchString(hostRegex, input)
	return
}

// Sets target.Target to the requried IP or cjdns path
func setTarget(data []string, usePath bool) (target Target, err error) {
	if len(data) == 0 {
		err = fmt.Errorf("Invalid target specified")
		return
	}
	input := data[0]
	if input != "" {

		if validIP(input) {
			target.Supplied = data[0]
			target.Target = padIPv6(net.ParseIP(input))
			return

		} else if validPath(input) && usePath {
			target.Target = input
			target.Supplied = data[0]
			return

		} else if validHost(input) {
			var ips []string
			ips, err = resolveHost(input)
			if err != nil {
				return
			}
			// Return the first result
			for _, addr := range ips {
				target.Target = addr
				target.Supplied = input
				return
			}

		} else {
			err = fmt.Errorf("Invalid IPv6 address, cjdns path, or hostname")
			return
		}
	}

	if usePath {
		err = fmt.Errorf("You must specify an IPv6 address, hostname or cjdns path")
		return
	}
	err = fmt.Errorf("You must specify an IPv6 address or hostname")
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
	println("ping <ipv6 address, hostname, or routing path>       sends a cjdns ping to the specified node")
	println("route <ipv6 address, hostname, or routing path>      prints out all routes to an IP or the IP to a route")
	println("traceroute <ipv6 address, hostname, or routing path> [-t timeout] performs a traceroute by pinging each known hop to the target on all known paths")
	println("ip <cjdns public key>                                converts a cjdns public key to the corresponding IPv6 address")
	println("host <ipv6 address or hostname>                      returns a list of all know IP address for the specified hostname or the hostname for an address")
	println("passgen                                              generates a random alphanumeric password between 15 and 50 characters in length")
	println("log [-l level] [-file file] [-line line]             prints cjdns log to stdout")
	println("peers                                                displays a list of currently connected peers")
	println("dump                                                 dumps the routing table to stdout")
	println("kill                                                 tells cjdns to gracefully exit")
	println("")
	println("Please use `cjdcmd --help` for a list of flags.")
}

// Returns a random alphanumeric string where length is <= max >= min
func randString(min, max int) string {
	r := myRand(min, max, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	return r
}

// Returns a random character from the specified string where length is <= max >= min
func myRand(min, max int, char string) string {

	var length int

	if min < max {
		length = min + rand.Intn(max-min)
	} else {
		length = min
	}

	buf := make([]byte, length)
	for i := 0; i < length; i++ {
		buf[i] = char[rand.Intn(len(char)-1)]
	}
	return string(buf)
}
