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
	target, err := setTarget(args, true)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	
	c := Connect()

	var tText string
	hostname, _ := resolveIP(target.Target)
	if hostname != "" {
		tText = target.Target + " (" + hostname + ")"
	} else {
		tText = target.Target
	}
	fmt.Printf("Showing all routes to %v\n", tText)

	table, err := c.NodeStore_dumpTable()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	table.SortByQuality()

	count := 0
	for _, v := range table {
		if v.IP.String() == target.Target || v.Path.String() == target.Target {
			if v.Link > 1 {
				fmt.Printf("IP: %v -- Version: %d -- Path: %s -- Link: %.0f\n", v.IP, v.Version, v.Path, v.Link)
				count++
			}
		}
	}
	fmt.Println("Found", count, "routes")
}
