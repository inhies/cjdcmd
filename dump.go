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

func dumpCmd(cmd *cobra.Command, args []string) {
	admin := Connect()
	table, err := admin.NodeStore_dumpTable()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	table.SortByQuality()
	k := 1
	for _, v := range table {
		if v.Link >= 1 {
			fmt.Fprintf(os.Stdout,
				"%03d IP: %-39v -- Version: %d -- Path: %s -- Link: %s\n",
				k, v.IP, v.Version, v.Path, v.Link)
			k++
		}
	}
}
