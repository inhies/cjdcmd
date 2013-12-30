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
)

type Routes []*Route

func (s Routes) Len() int      { return len(s) }
func (s Routes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByPath struct{ Routes }

func (s ByPath) Less(i, j int) bool { return s.Routes[i].RawPath < s.Routes[j].RawPath }

//Sorts with highest quality link at the top
type ByQuality struct{ Routes }

func (s ByQuality) Less(i, j int) bool { return s.Routes[i].RawLink > s.Routes[j].RawLink }

// TODO(inhies): Make the output nicely formatted
func doTraceroute(user *admin.Conn, target Target) {
	table, err := user.NodeStore_dumpTable()
	if err != nil {
		fmt.Println(err)
		return
	}
	usingPath := false
	var tText string
	if validIP(target.Supplied) {
		hostname, _ := resolveIP(target.Target)
		if hostname != "" {
			tText = target.Supplied + " (" + hostname + ")"
		} else {
			tText = target.Supplied
		}
		// If we were given a path, resolve the IP
	} else if validPath(target.Supplied) {
		usingPath = true
		tText = target.Supplied
		//table := getTable(globalData.User)
		for _, v := range table {
			if v.Path.String() == target.Supplied {
				// We have the IP now
				tText = target.Supplied + " (" + v.IP.String() + ")"

				// Try to get the hostname
				hostname, _ := resolveIP(v.IP.String())
				if hostname != "" {
					tText = target.Supplied + " (" + v.IP.String() + " (" + hostname + "))"
				}
			}
		}
		// We were given a hostname, everything is already done for us!
	} else if validHost(target.Supplied) {
		tText = target.Supplied + " (" + target.Target + ")"
	}

	fmt.Println("Finding all routes to", tText)

	count := 0
	for i := range table {
		if usingPath {
			if table[i].Path.String() != target.Supplied {
				continue
			}
		} else {
			if table[i].IP.String() != target.Target {
				continue
			}
		}

		if table[i].Link < 1 {
			continue
		}

		response := table.Hops(*table[i].Path)
		if err != nil {
			fmt.Println("Error:", err)
		}
		response.SortByPath()
		count++
		fmt.Printf("\nRoute #%d to target: %v\n", count, table[i].Path)
		for y, p := range response {
			hostname, _ := resolveIP(p.IP.String())
			var IP string
			if hostname != "" {
				IP = p.IP.String() + " (" + hostname + ")"
			} else {
				IP = p.IP.String()
			}
			fmt.Printf("IP: %v -- Version: %d -- Path: %s -- Link: %s -- Time:", IP, p.Version, p.Path, p.Link)
			if y == 0 {
				fmt.Printf(" Skipping ourself\n")
				continue
			}
			for x := 1; x <= 3; x++ {
				tRoute := &Ping{}
				tRoute.Target = p.Path.String()
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
			fmt.Println("")
		}
	}
	fmt.Println("Found", count, "routes")
}

/*
func getHops(table admin.Routes, fullPath admin.Path) (output admin.Routes, err error) {
	for i := range table {
		candPath := table[i].Path

		g := 64 - uint64(math.Log2(float64(candPath)))
		h := uint64(uint64(0xffffffffffffffff) >> g)

		if h&uint64(fullPath) == h&uint64(candPath) {
			output = append(output, table[i])
		}
	}
	return
}
*/
