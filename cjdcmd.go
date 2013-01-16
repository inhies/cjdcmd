// Package tool is a tool for using cjdns
package main

import (
	"encoding/binary"
	"encoding/hex"
	"flag"
	"fmt"
	"github.com/inhies/go-cjdns/admin"
	"github.com/inhies/go-cjdns/config"
	"math"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
)

const (
	Version = "0.2"

	defaultPingTimeout = 5000 //5 seconds
	defaultPingCount   = 0

	defaultLogLevel    = "DEBUG"
	defaultLogFile     = ""
	defaultLogFileLine = 0

	defaultFile = "/etc/cjdroute.conf"

	pingCmd     = "ping"
	logCmd      = "log"
	traceCmd    = "traceroute"
	peerCmd     = "peers"
	dumpCmd     = "dump"
	routeCmd    = "route"
	killCmd     = "kill"
	versionsCmd = "versions"

	magicalLinkConstant = 5366870.0
)

var (
	PingTimeout int
	PingCount   int

	LogLevel    string
	LogFile     string
	LogFileLine int

	fs *flag.FlagSet

	File string
)

type Ping struct {
	IP, Version, Response, Error                 string
	Failed, Percent, Sent, Success               float64
	CTime, TTime, TTime2, TMin, TAvg, TMax, TDev float64
}
type Route struct {
	IP      string
	Path    string
	RawPath uint64
	Link    float64
	RawLink int64
	Version int64
}
type Routes []*Route

func (s Routes) Len() int      { return len(s) }
func (s Routes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByPath struct{ Routes }

func (s ByPath) Less(i, j int) bool { return s.Routes[i].RawPath < s.Routes[j].RawPath }

func init() {

	fs = flag.NewFlagSet("cjdcmd", flag.ExitOnError)
	const (
		usagePingTimeout = "[ping][traceroute] specify the time in milliseconds cjdns should wait for a response"
		usagePingCount   = "[ping][traceroute] specify the number of packets to send"

		usageLogLevel    = "[log] specify the logging level to use"
		usageLogFile     = "[log] specify the cjdns source file you wish to see log output from"
		usageLogFileLine = "[log] specify the cjdns source file line to log"

		usageFile = "[all] the cjdroute.conf configuration file to use, edit, or view"
	)
	fs.StringVar(&File, "file", defaultFile, usageFile)
	fs.StringVar(&File, "f", defaultFile, usageFile+" (shorthand)")

	fs.IntVar(&PingTimeout, "timeout", defaultPingTimeout, usagePingTimeout)
	fs.IntVar(&PingTimeout, "t", defaultPingTimeout, usagePingTimeout+" (shorthand)")

	fs.IntVar(&PingCount, "count", defaultPingCount, usagePingCount)
	fs.IntVar(&PingCount, "c", defaultPingCount, usagePingCount+" (shorthand)")

	fs.StringVar(&LogLevel, "level", defaultLogLevel, usageLogLevel)
	fs.StringVar(&LogLevel, "l", defaultLogLevel, usageLogLevel+" (shorthand)")

	fs.StringVar(&LogFile, "logfile", defaultLogFile, usageLogFile)
	fs.IntVar(&LogFileLine, "line", defaultLogFileLine, usageLogFileLine)

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
	println("ping <ipv6 address or cjdns routing path>  sends a cjdns ping to the specified node")
	println("route <ipv6 address or cjdns routing path> prints out all routes to an IP or the IP to a route")
	println("traceroute <ipv6 addressh> [-t timeout]    performs a traceroute by pinging each known hop to the target on all known paths")
	println("log [-l level] [-file file] [-line line]   prints cjdns log to stdout")
	println("peers                                      displays a list of currently connected peers")
	println("dump                                       dumps the routing table to stdout")
	println("kill                                       tells cjdns to gracefully exit")
	println("")
	println("Please use `cjdcmd --help` for a list of flags.")
}

func outputPing(Ping *Ping) {

	if Ping.Success > 0 {
		Ping.TAvg = Ping.TTime / Ping.Success
	}
	Ping.TTime2 /= Ping.Success

	if Ping.Success > 0 {
		Ping.TDev = math.Sqrt(Ping.TTime2 - Ping.TAvg*Ping.TAvg)
	}
	Ping.Percent = (Ping.Failed / Ping.Sent) * 100

	fmt.Println("\n---", Ping.IP, "ping statistics ---")
	fmt.Printf("%v packets transmitted, %v received, %.2f%% packet loss, time %vms\n", Ping.Sent, Ping.Success, Ping.Percent, Ping.TTime)
	fmt.Printf("rtt min/avg/max/mdev = %.3f/%.3f/%.3f/%.3f ms\n", Ping.TMin, Ping.TAvg, Ping.TMax, Ping.TDev)
	fmt.Printf("Target is using cjdns version %v\n", Ping.Version)
}

func main() {
	//Define the flags, and parse them
	//Clearly a hack but it works for now
	//TODO(inhies): Re-implement flag parsing so flags can have multiple meanings based on the base command (ping, route, etc)
	if len(os.Args) <= 1 {
		usage()
		return
	} else if len(os.Args) == 2 {
		if string(os.Args[1]) == "--help" {
			fs.PrintDefaults()
			return
		}
	} else {
		fs.Parse(os.Args[2:])
	}

	//TODO(inhies): check argv[0] for trailing commands.
	//For example, to run ctraceroute:
	//ln -s /path/to/cjdcmd /usr/bin/ctraceroute like things
	command := os.Args[1]

	//read the config
	// TODO: check ./cjdroute.conf /etc/cjdroute.conf ~/cjdroute.conf ~/cjdns/cjdroute.conf maybe ~/cjdns/build/cjdroute.conf
	conf, err := config.LoadMinConfig(File)

	if err != nil || len(conf.Admin.Password) == 0 {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Attempting to connect to cjdns...")
	user, err := admin.Connect(conf.Admin.Bind, conf.Admin.Password) //conf.Admin.Bind
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

	println("Connected")
	defer user.Conn.Close()
	arguments := fs.Args()
	data := arguments[fs.NFlag()-fs.NFlag():]

	//Setup variables now so that if the program is killed we can still finish what we're doing
	ping := &Ping{}
	var loggingStreamID string

	// capture ctrl+c 
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Printf("\n")
			if command == "log" {
				//unsubscribe from logging
				_, err := admin.AdminLog_unsubscribe(user, loggingStreamID)
				if err != nil {
					fmt.Printf("%v\n", err)
					return
				}
			}
			if command == "ping" {
				//stop pinging and print results
				outputPing(ping)
			}
			//close all the channels
			for _, c := range user.Channels {
				close(c)
			}
			user.Conn.Close()
			return
		}
	}()

	switch command {

	case versionsCmd:
		// TODO: ping all nodes in the routing table and get their versions
		// git log -1 --date=iso --pretty=format:"%ad" <hash>
	case killCmd:
		_, err := admin.Core_exit(user)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		alive := true
		for ; alive; alive, _ = admin.SendPing(user, 1000) {
			runtime.Gosched() //play nice
		}
		println("cjdns is shutting down...")

	case dumpCmd:
		// TODO: add flag to show zero link quality routes, by default hide them
		table := getTable(user)
		k := 1
		for _, v := range table {
			if v.Link >= 1 {
				fmt.Printf("%d IP: %v -- Version: %d -- Path: %s -- Link: %.0f\n", k, v.IP, v.Version, v.Path, v.Link)
				k++
			}
		}

	case traceCmd:
		var target string
		if len(data) > 0 {
			if strings.Count(data[0], ":") > 1 {
				target = padIPv6(net.ParseIP(data[0]))
				if len(target) != 39 {
					fmt.Println("Invalid IPv6 address")
					return
				}
			} else {
				fmt.Println("Invalid IPv6 address")
				return
			}
		} else {
			fmt.Println("You must specify an IPv6 address")
			return
		}
		table := getTable(user)
		fmt.Println("Finding all routes to", target)

		count := 1
		for i := range table {

			if table[i].IP != target {
				continue
			}

			if table[i].Link < 1 {
				continue
			}

			response, err := getHops(table, table[i].RawPath)
			if err != nil {
				fmt.Println(err)
			}

			sort.Sort(ByPath{response})

			fmt.Printf("\nRoute #%d to target\n", count)
			for y, p := range response {

				fmt.Printf("IP: %v -- Version: %d -- Path: %s -- Link: %.0f -- Time:", p.IP, p.Version, p.Path, p.Link)
				if y == 0 {
					fmt.Printf(" Skipping ourself\n")
					continue
				}
				for x := 1; x <= 3; x++ {
					tRoute := &Ping{}
					tRoute.IP = p.Path
					err := pingNode(user, tRoute)
					if err != nil {
						fmt.Println("Error:", err)
						return
					}
					if tRoute.Error == "timeout" {
						fmt.Printf("   *  ")
					} else {
						fmt.Printf(" %vms", tRoute.TTime)
					}
				}
				println("")
			}
			count++
		}
	case "memory":
		println("Bye bye cjdns! This command causes a crash...")
		response, err := admin.Memory(user)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		fmt.Println(response)
	case routeCmd:
		var target string
		if len(data) > 0 {
			if strings.Count(data[0], ":") > 1 {
				target = padIPv6(net.ParseIP(data[0]))
				if len(target) != 39 {
					fmt.Println("Invalid IPv6 address")
					return
				}
			} else if strings.Count(data[0], ".") == 3 {
				target = data[0]
				if len(target) != 19 {
					fmt.Println("Invalid cjdns path")
					return
				}
			} else {
				fmt.Println("Invalid IPv6 address or cjdns path")
				return
			}
		} else {
			fmt.Println("You must specify an IPv6 address or cjdns path")
			return
		}
		table := getTable(user)
		for _, v := range table {
			if v.IP == target || v.Path == target {
				if v.Link > 1 {
					fmt.Printf("IP: %v -- Version: %d -- Path: %s -- Link: %.0f\n", v.IP, v.Version, v.Path, v.Link)
				}
			}
		}
		return
	case pingCmd:
		// TODO: allow input of IP, hex path with and without dots and leading zeros, and binary path
		// TODO: allow pinging of entire routing table
		if len(data) > 0 {
			if strings.Count(data[0], ":") > 1 {
				ping.IP = padIPv6(net.ParseIP(data[0]))
				if len(ping.IP) != 39 {
					fmt.Println("Invalid IPv6 address")
					return
				}
			} else if strings.Count(data[0], ".") == 3 {
				ping.IP = data[0]
				if len(ping.IP) != 19 {
					fmt.Println("Invalid cjdns path")
					return
				}
			} else {
				fmt.Println("Invalid IPv6 address or cjdns path")
				return
			}
		} else {
			fmt.Println("You must specify an IPv6 address or cjdns path")
			return
		}

		if PingCount != defaultPingCount {
			// ping only as much as the user asked for
			for i := 1; i <= PingCount; i++ {
				err := pingNode(user, ping)
				if err != nil {
					fmt.Println(err)
					return
				}
				println(ping.Response)
			}
		} else {
			// ping until we're told otherwise
			for {
				err := pingNode(user, ping)
				if err != nil {
					fmt.Println(err)
					return
				}
				println(ping.Response)
			}
		}
		outputPing(ping)

	case logCmd:
		var response chan map[string]interface{}
		response, loggingStreamID, err = admin.AdminLog_subscribe(user, LogFile, LogLevel, LogFileLine)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
		format := "%d %d %s %s:%d %s\n" // TODO: add user formatted output
		counter := 1
		for {
			input, ok := <-response
			if !ok {
				break
			}
			fmt.Printf(format, counter, input["time"], input["level"], input["file"], input["line"], input["message"])
			counter++
		}

	case peerCmd:
		peers := make([]*Route, 0)
		table := getTable(user)

		fmt.Println("Finding all connected peers")

		for i := range table {

			if table[i].Link < 1 {
				continue
			}
			if table[i].RawPath == 1 {
				continue
			}
			response, err := getHops(table, table[i].RawPath)
			if err != nil {
				fmt.Println(err)
			}

			sort.Sort(ByPath{response})

			var peer *Route
			if len(response) > 1 {
				peer = response[1]
			} else {
				peer = response[0]
			}

			found := false
			for _, p := range peers {
				if p == peer {
					found = true
					break
				}
			}

			if !found {
				peers = append(peers, peer)
			}
		}
		for _, p := range peers {
			fmt.Printf("IP: %v -- Path: %s -- Link: %.0f\n", p.IP, p.Path, p.Link)
		}
	default:
		fmt.Println("Invalid command", command)
		usage()
		return
	}
}

func getHops(table []*Route, fullPath uint64) (output []*Route, err error) {
	for i := range table {
		candPath := table[i].RawPath

		g := 64 - uint64(math.Log2(float64(candPath)))
		h := uint64(uint64(0xffffffffffffffff) >> g)

		if h&fullPath == h&candPath {
			output = append(output, table[i])
		}
	}
	return
}

/*
case "test":
		table := getTable(user)
		host := "fcf1:b5d5:d0b4:c390:9db2:3f5e:d2d2:bff2"

		for _, v := range table {
			if v.IP == host {
				sPath1 := strings.Replace(v.RawPath, ".", "", -1)
				bPath1, _ := hex.DecodeString(sPath1)
				path := binary.BigEndian.Uint64(bPath1)

				result, err := subPath(table, path)
				if err != nil {
					fmt.Println(err)
				}
				if result != 0 {
					println("found  a path")
				}
			}
		}

	default:
		fmt.Println("Invalid command", command)
		usage()
		return
	}
}
*/

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
	table = make([]*Route, 0)
	page := 0
	var more int64
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

// Pings a node and generates statistics
func pingNode(user *admin.Admin, ping *Ping) (err error) {
	response, err := admin.RouterModule_pingNode(user, ping.IP, PingTimeout)

	if err != nil {
		return
	}

	ping.Sent++
	if response.Error == "" {
		if response.Result == "timeout" {
			ping.Response = fmt.Sprintf("Timeout from %v after %vms", ping.IP, response.Time)
			ping.Error = "timeout"
			ping.Failed++
		} else {
			ping.Response = fmt.Sprintf("Reply from %v %vms", ping.IP, response.Time)
			ping.Success++
			ping.CTime = float64(response.Time)
			ping.TTime += ping.CTime
			ping.TTime2 += ping.CTime * ping.CTime
			if ping.TMin == 0 {
				ping.TMin = ping.CTime
			}
			if ping.CTime > ping.TMax {
				ping.TMax = ping.CTime
			}
			if ping.CTime < ping.TMin {
				ping.TMin = ping.CTime
			}

			if ping.Version == "" {
				ping.Version = response.Version
			}
			if ping.Version != response.Version {
				//not likely we'll see this happen but it doesnt hurt to be prepared
				println("Host is sending back mismatched versions")
			}
		}
	} else {
		ping.Failed++
		err = fmt.Errorf(response.Error)
		ping.Error = response.Error
		return
	}
	return
}
