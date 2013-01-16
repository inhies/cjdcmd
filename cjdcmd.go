// Package tool is a tool for using cjdns
package main

import (
	"flag"
	"fmt"
	"github.com/inhies/go-cjdns/admin"
	"github.com/inhies/go-cjdns/config"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sort"
)

const (
	Version = "0.2.2"

	defaultPingTimeout = 5000 //5 seconds
	defaultPingCount   = 0

	defaultLogLevel    = "DEBUG"
	defaultLogFile     = ""
	defaultLogFileLine = 0

	defaultFile = "/etc/cjdroute.conf"

	pingCmd    = "ping"
	logCmd     = "log"
	traceCmd   = "traceroute"
	peerCmd    = "peers"
	dumpCmd    = "dump"
	routeCmd   = "route"
	killCmd    = "kill"
	versionCmd = "version"

	magicalLinkConstant = 5366870.0

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

	File string
)

type Route struct {
	IP      string
	Path    string
	RawPath uint64
	Link    float64
	RawLink int64
	Version int64
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

	case traceCmd:
		target, err := setTarget(data, false)
		if err != nil {
			fmt.Println(err)
			return
		}
		doTraceroute(user, target)

	case routeCmd:
		target, err := setTarget(data, true)
		if err != nil {
			fmt.Println(err)
			return
		}
		table := getTable(user)
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
		ping.Target = target
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
	case versionCmd:
		// TODO(inhies): Ping a specific node and return it's cjdns version, or
		// ping all nodes in the routing table and get their versions
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
		sort.Sort(ByQuality{table})
		k := 1
		for _, v := range table {
			if v.Link >= 1 {
				fmt.Printf("%d IP: %v -- Version: %d -- Path: %s -- Link: %.0f\n", k, v.IP, v.Version, v.Path, v.Link)
				k++
			}
		}
	case "memory":
		println("Bye bye cjdns! This command causes a crash. Keep trying and maybe one day cjd will fix it :)")
		response, err := admin.Memory(user)
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
