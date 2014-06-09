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

func routeCmd(cmd *cobra.Command, args []string) {
	c := Connect()

	table, err := c.NodeStore_dumpTable()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	table.SortByQuality()

	for _, arg := range args {
		hostname, ip, _ := resolve(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not resolve %s, %s\n", arg, err)
			continue
		}

		fmt.Fprintf(os.Stdout, "Showing all routes to %s (%s)\n", hostname, ip)

		count := 0
		for _, r := range table {
			if ip.Equal(*r.IP) {
				if r.Link > 1 {
					fmt.Fprintf(os.Stdout, "Path: %s -- Link: %d\n", r.Path, r.Link)
					count++
				}
			}
		}
		fmt.Fprintf(os.Stdout, "Found %d routes\n\n", count)
	}
}
