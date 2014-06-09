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
	"github.com/spf13/cobra"
	"os"
)

func peersCmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		c := Connect()
		stats, err := c.InterfaceController_peerStats()
		if err != nil {
			fmt.Println("Error getting local peers,", err)
		}

		var addr string
		for _, node := range stats {
			addr = node.PublicKey.IP().String()
			fmt.Fprintf(os.Stdout, "Incoming: %t | IP: %-39s -- Path: %s\n", node.IsIncoming, addr, node.SwitchLabel)
		}
		return
	}

	if len(args) > 1 {
		cmd.Usage()
		os.Exit(1)
	}

	_, ip, err := resolve(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not resolve "+args[0]+".")
		os.Exit(1)
	}
	
	c := Connect()
	table, err := c.NodeStore_dumpTable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get routing table:", err)
		os.Exit(1)
	}

	peers := table.Peers(ip)

	if len(peers) == 0 {
		fmt.Fprintln(os.Stderr, "no peers found in local routing table")
		os.Exit(1)
	}
	for _, p := range peers {
		host, _, _ := resolve(p.IP.String())
		fmt.Println("\t ", p.IP, host)
		//fmt.Printf("IP: %v -- Path: %s -- Link: %.0f\n", tText, node.Path, node.Link)
	}
}
