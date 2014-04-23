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
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
)

var (
	// BUG(emery): oft useless instantiation
	rCache = make(map[string]string)
	rMutex = new(sync.RWMutex)
)

func resolve(host string) (hostname string, ip net.IP, err error) {
	if ipRegex.MatchString(host) {
		ip = net.ParseIP(host)
		addr := ip.String()

		var ok bool
		rMutex.RLock()
		hostname, ok = rCache[addr]
		rMutex.RUnlock()
		if ok {
			return
		}
		rMutex.Lock()
		defer rMutex.Unlock()
		hostname, ok = rCache[addr]
		if ok {
			return
		}

		if hostnames, err := net.LookupAddr(addr); err == nil {
			for _, h := range hostnames {
				hostname = hostname + h + " "
			}
		} else if Verbose {
			fmt.Fprintf(os.Stderr, "failed to resolve %s, %s\n", addr, err)
		}

		if len(hostname) > 0 {
			rCache[addr] = hostname
		}
		return
	} else {
		hostname = host
		var addrs []string
		if addrs, err = net.LookupHost(host); err != nil {
			return
		}

		for _, addr := range addrs {
			if ipRegex.MatchString(addr) {
				ip = net.ParseIP(addr)
				return
			}
		}
		err = errors.New("no fc::/8 address found")
	}
	return
}
