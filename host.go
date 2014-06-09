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
)

func hostCmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		os.Exit(1)
	}

	var table admin.Routes

	for _, arg := range args {

		if ipRegex.MatchString(arg) {
			hostname, err := resolveIP(arg)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			fmt.Printf("%v\n", hostname)
		} else if pathRegex.MatchString(arg) {
			if len(table) == 0 {
				c := Connect()
				t, err := c.NodeStore_dumpTable()
				if err != nil {
					fmt.Fprintln(os.Stderr, "Error finding path in table:", err)
					os.Exit(1)
				}
				table = t
			}
		tableLoop:
			for _, node := range table {
				if node.Path.String() == arg {
					fmt.Fprintf(os.Stdout, "%v has IPv6 address %v\n", arg, node.IP)
					break tableLoop
				}
			}

		} else if hostRegex.MatchString(arg) {
			ips, err := resolveHost(arg)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error:", err)
				os.Exit(1)
			}
			for _, addr := range ips {
				fmt.Fprintf(os.Stdout, "%v has IPv6 address %v\n", arg, addr)
			}
		} else {
			fmt.Fprintln(os.Stderr, "Invalid hostname or IPv6 address specified")
			os.Exit(1)
		}
	}
}
