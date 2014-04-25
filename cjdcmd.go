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
	"github.com/inhies/go-cjdns/admin"
	"github.com/spf13/cobra"
	"os"
	"regexp"
)

const Version = "0.6.0"

var (
	ipRegex   = regexp.MustCompile("^fc[a-f0-9]{1,2}:([a-f0-9]{0,4}:){2,6}[a-f0-9]{1,4}$")
	pathRegex = regexp.MustCompile("([0-9a-f]{4}\\.){3}[0-9a-f]{4}")
	hostRegex = regexp.MustCompile("^([a-zA-Z0-9]([a-zA-Z0-9\\-\\.]{0,}[a-zA-Z0-9]))$")
)

var (
	ConfFileIn, ConfFileOut   string
	AdminFileIn, AdminFileOut string

	NmapOutput bool
	Verbose    bool
)

var (
	rootCmd = &cobra.Command{Use: "cjdcmd"}

	PingCmd = &cobra.Command{
		Use:   "ping <IPv6/DNS>",
		Short: "Preforms a cjdns ping to a specified address.",
		Run:   pingCmd,
	}

	RouteCmd = &cobra.Command{
		Use:   "route <IPv6/DNS/Path>",
		Short: "Prints all routes to a specific node",
		Run:   routeCmd,
	}

	TracerouteCmd = &cobra.Command{
		Use:   "traceroute <IPv6/DNS/Path>",
		Short: "Performs a traceroute on a specific node by pinging each known hop to the target on all known paths",
		Run:   tracerouteCmd,
	}

	PubKeyToIPCmd = &cobra.Command{
		Use:   "ip <cjdns public key>",
		Short: "Converts a cjdns public key to its corresponding IPv6 address.",
		Run:   pubKeyToIPCmd,
	}

	PeersCmd = &cobra.Command{
		Use:   "peers [<IPv6/DNS/Path>]",
		Short: "Displays a list of currently connected peers for a node, if no node is specified your peers are shown.",
		Run:   peersCmd,
	}

	HostCmd = &cobra.Command{
		Use:   "host <IPv6/DNS>",
		Short: "Returns a list of all known IP addresses for a specified hostname or the hostname for an address.",
		Run:   hostCmd,
	}

	CjdnsAdminCmd = &cobra.Command{
		Use:   "cjdnsadmin <-file /path/to/cjdroute.conf>",
		Short: "Generates a .cjdnsadmin file in your home diectory using the specified cjdroute.conf as input",
		Run:   cjdnsAdminCmd,
	}

	AddPeerCmd = &cobra.Command{
		Use:   "addpeer '<json peer details>'",
		Short: "Adds the peer details to your config file",
		Long:  "You must enter the peering details surrounded by single qoutes '<peer details>'",
		Run:   addPeerCmd,
	}

	AddPasswordCmd = &cobra.Command{
		Use:   "addpass [password]",
		Short: "Adds the password to your config file, or generates one and then adds that",
		Run:   addPasswordCmd,
	}

	ListPasswordCmd = &cobra.Command{
		Use:   "listpass",
		Short: "ALPHA FEATURE - List currently loaded peering passwords.",
		Run:   listPasswordCmd,
	}

	CleanConfigCmd = &cobra.Command{
		Use:   "cleanconfig [-file] [-outfile]",
		Short: "Strips all comments from the config file and saves it at outfile",
		Run:   cleanConfigCmd,
	}

	LogCmd = &cobra.Command{
		Use:   "log [--level level] [--file file] [--line]",
		Short: "Prints cjdns logs to stdout",
		Run:   logCmd,
	}

	PassGenCmd = &cobra.Command{
		Use:   "passgen [prefix]",
		Short: "Generates a random alphanumeric password between 15 and 50 characters. If you provide [prefix], it will be prepended. This is to help you keep track of your peering passwords",
		Run:   passGenCmd,
	}

	DumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dumps the entire routing table to stdout.",
		Run:   dumpCmd,
	}

	KillCmd = &cobra.Command{
		Use:   "kill",
		Short: "Gracefully kills cjdns",
	}

	MemoryCmd = &cobra.Command{
		Use:   "memory",
		Short: "Returns the bytes of memory allocated by the router",
		Run:   memoryCmd,
	}
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")

	TracerouteCmd.Flags().BoolVarP(&NmapOutput, "nmap", "x", false, "format output as nmap XML")

	rootCmd.AddCommand(
		PingCmd,
		RouteCmd,
		TracerouteCmd,
		PubKeyToIPCmd,
		PeersCmd,
		HostCmd,
		CjdnsAdminCmd,
		AddPeerCmd,
		AddPasswordCmd,
		ListPasswordCmd,
		CleanConfigCmd,
		LogCmd,
		PassGenCmd,
		DumpCmd,
		KillCmd,
		MemoryCmd,
	)
}

func main() { rootCmd.Execute() }

// Connect connects to
func Connect() *admin.Conn {
	admin, err := admin.Connect(nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not connect to cjdns:", err)
		os.Exit(1)
	}
	return admin
}
