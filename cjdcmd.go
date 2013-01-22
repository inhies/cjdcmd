// Package tool is a tool for using cjdns
package main

import (
	//"encoding/hex"
	"flag"
	"fmt"
	"github.com/inhies/go-cjdns/admin"

	"math/rand"

	"os"
	"os/signal"
	"runtime"
	"sort"
	"time"
)

const (
	Version = "0.2.3"

	defaultPingTimeout = 5000 //5 seconds
	defaultPingCount   = 0

	defaultLogLevel    = "DEBUG"
	defaultLogFile     = ""
	defaultLogFileLine = 0

	defaultFile      = "/etc/cjdroute.conf"
	defaultPass      = ""
	defaultAdminBind = "127.0.0.1:11234"

	pingCmd       = "ping"
	logCmd        = "log"
	traceCmd      = "traceroute"
	peerCmd       = "peers"
	dumpCmd       = "dump"
	routeCmd      = "route"
	killCmd       = "kill"
	versionCmd    = "version"
	pubKeyToIPcmd = "ip"
	passGenCmd    = "passgen"

	magicalLinkConstant = 5366870.0 //Determined by cjd way back in the dark ages.

	ipRegex   = "^fc[a-f0-9]{1,2}:([a-f0-9]{0,4}:){2,6}[a-f0-9]{1,4}$"
	pathRegex = "([0-9a-f]{4}\\.){3}[0-9a-f]{4}"
	hostRegex = "^([a-zA-Z0-9]([a-zA-Z0-9\\-\\.]{0,}[a-zA-Z0-9]))$"
)

var (
	PingTimeout int
	PingCount   int

	LogLevel    string
	LogFile     string
	LogFileLine int

	fs *flag.FlagSet

	File          string
	AdminPassword string
	AdminBind     string
)

type Route struct {
	IP      string
	Path    string
	RawPath uint64
	Link    float64
	RawLink int64
	Version int64
}

type Data struct {
	User            *admin.Admin
	LoggingStreamID string
}

func init() {

	fs = flag.NewFlagSet("cjdcmd", flag.ExitOnError)
	const (
		usagePingTimeout = "[ping][traceroute] specify the time in milliseconds cjdns should wait for a response"
		usagePingCount   = "[ping][traceroute] specify the number of packets to send"

		usageLogLevel    = "[log] specify the logging level to use"
		usageLogFile     = "[log] specify the cjdns source file you wish to see log output from"
		usageLogFileLine = "[log] specify the cjdns source file line to log"

		usageFile = "[all] the cjdroute.conf configuration file to use, edit, or view"

		usagePass = "[all] specify the admin password"
	)
	fs.StringVar(&File, "file", defaultFile, usageFile)
	fs.StringVar(&File, "f", defaultFile, usageFile+" (shorthand)")

	fs.IntVar(&PingTimeout, "timeout", defaultPingTimeout, usagePingTimeout)
	fs.IntVar(&PingTimeout, "t", defaultPingTimeout, usagePingTimeout+" (shorthand)")

	fs.IntVar(&PingCount, "count", defaultPingCount, usagePingCount)
	fs.IntVar(&PingCount, "c", defaultPingCount, usagePingCount+" (shorthand)")

	fs.StringVar(&LogLevel, "level", defaultLogLevel, usageLogLevel)
	fs.StringVar(&LogLevel, "l", defaultLogLevel, usageLogLevel+" (shorthand)")

	fs.StringVar(&AdminPassword, "pass", defaultPass, usagePass)
	fs.StringVar(&AdminPassword, "p", defaultPass, usagePass+" (shorthand)")

	fs.StringVar(&LogFile, "logfile", defaultLogFile, usageLogFile)
	fs.IntVar(&LogFileLine, "line", defaultLogFileLine, usageLogFileLine)

	rand.Seed(time.Now().UTC().UnixNano())
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

	arguments := fs.Args()
	data := arguments[fs.NFlag()-fs.NFlag():]

	//Setup variables now so that if the program is killed we can still finish what we're doing
	ping := &Ping{}

	var globalData Data
	// capture ctrl+c 
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for _ = range c {
			fmt.Printf("\n")
			if command == "log" {
				//unsubscribe from logging
				_, err := admin.AdminLog_unsubscribe(globalData.User, globalData.LoggingStreamID)
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
			for _, c := range globalData.User.Channels {
				close(c)
			}
			globalData.User.Conn.Close()
			return
		}
	}()

	switch command {
	case passGenCmd:
		// Prints a random alphanumberic password between 15 and 50 characters long
		// TODO(inies): Make more better
		println(randString(15, 50))
	case pubKeyToIPcmd:
		var ip []byte
		if len(data) > 0 {
			if len(data[0]) == 52 || len(data[0]) == 54 {
				ip = []byte(data[0])
			} else {
				println("Invalid public key")
				return
			}
		} else {
			println("Invalid public key")
			return
		}
		parsed, err := admin.PubKeyToIP(ip)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("%v\n", parsed)
	case traceCmd:
		user, err := connect()
		if err != nil {
			fmt.Println(err)
			return
		}
		globalData.User = user
		target, err := setTarget(data, false)
		if err != nil {
			fmt.Println(err)
			return
		}
		doTraceroute(globalData.User, target)

	case routeCmd:

		target, err := setTarget(data, true)
		if err != nil {
			fmt.Println(err)
			return
		}
		user, err := connect()
		if err != nil {
			fmt.Println(err)
			return
		}
		globalData.User = user
		table := getTable(globalData.User)
		sort.Sort(ByQuality{table})
		count := 0
		for _, v := range table {
			if v.IP == target || v.Path == target {
				if v.Link > 1 {
					fmt.Printf("IP: %v -- Version: %d -- Path: %s -- Link: %.0f\n", v.IP, v.Version, v.Path, v.Link)
					count++
				}
			}
		}
		fmt.Println("Found", count, "routes")

	case pingCmd:
		// TODO: allow input of IP, hex path with and without dots and leading zeros, and binary path
		// TODO: allow pinging of entire routing table
		target, err := setTarget(data, true)
		if err != nil {
			fmt.Println(err)
			return
		}
		user, err := connect()
		if err != nil {
			fmt.Println(err)
			return
		}
		globalData.User = user
		ping.Target = target
		if PingCount != defaultPingCount {
			// ping only as much as the user asked for
			for i := 1; i <= PingCount; i++ {
				start := time.Duration(time.Now().UTC().UnixNano())
				err := pingNode(globalData.User, ping)
				if err != nil {
					if err.Error() != "Socket closed" {
						fmt.Println(err)
					}
					return
				}
				println(ping.Response)
				// Send 1 ping per second
				now := time.Duration(time.Now().UTC().UnixNano())
				time.Sleep(start + time.Second - now)

			}
		} else {
			// ping until we're told otherwise
			for {
				start := time.Duration(time.Now().UTC().UnixNano())
				err := pingNode(globalData.User, ping)
				if err != nil {
					if err.Error() != "Socket closed" {
						fmt.Println(err)
					}
					return
				}
				println(ping.Response)
				// Send 1 ping per second
				now := time.Duration(time.Now().UTC().UnixNano())
				time.Sleep(start + time.Second - now)

			}
		}
		outputPing(ping)

	case logCmd:
		user, err := connect()
		if err != nil {
			fmt.Println(err)
			return
		}
		globalData.User = user
		var response chan map[string]interface{}
		response, globalData.LoggingStreamID, err = admin.AdminLog_subscribe(globalData.User, LogFile, LogLevel, LogFileLine)
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
		user, err := connect()
		if err != nil {
			fmt.Println(err)
			return
		}
		globalData.User = user
		peers := make([]*Route, 0)
		table := getTable(globalData.User)
		sort.Sort(ByQuality{table})
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
	case versionCmd:
		// TODO(inhies): Ping a specific node and return it's cjdns version, or
		// ping all nodes in the routing table and get their versions
		// git log -1 --date=iso --pretty=format:"%ad" <hash>

	case killCmd:
		user, err := connect()
		if err != nil {
			fmt.Println(err)
			return
		}
		globalData.User = user
		_, err = admin.Core_exit(globalData.User)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		alive := true
		for ; alive; alive, _ = admin.SendPing(globalData.User, 1000) {
			runtime.Gosched() //play nice
		}
		println("cjdns is shutting down...")

	case dumpCmd:
		user, err := connect()
		if err != nil {
			fmt.Println(err)
			return
		}
		globalData.User = user
		// TODO: add flag to show zero link quality routes, by default hide them
		table := getTable(globalData.User)
		sort.Sort(ByQuality{table})
		k := 1
		for _, v := range table {
			if v.Link >= 1 {
				fmt.Printf("%d IP: %v -- Version: %d -- Path: %s -- Link: %.0f\n", k, v.IP, v.Version, v.Path, v.Link)
				k++
			}
		}
	case "memory":
		user, err := connect()
		if err != nil {
			fmt.Println(err)
			return
		}
		globalData.User = user
		println("Bye bye cjdns! This command causes a crash. Keep trying and maybe one day cjd will fix it :)")
		response, err := admin.Memory(globalData.User)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		fmt.Println(response)
	default:
		fmt.Println("Invalid command", command)
		usage()
	}
}
