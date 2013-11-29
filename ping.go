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
	"math"
)

type Ping struct {
	Target, Domain, Version, Response, Error     string
	Failed, Percent, Sent, Success               float64
	CTime, TTime, TTime2, TMin, TAvg, TMax, TDev float64
}

// Pings a node and generates statistics
func pingNode(user *admin.Conn, ping *Ping) (err error) {
	response, err := user.RouterModule_pingNode(ping.Target, PingTimeout)

	if err != nil {
		return
	}

	ping.Sent++
	if response.Error == "" {
		if response.Result == "timeout" {
			ping.Response = fmt.Sprintf("Timeout from %v after %vms", ping.Target, response.Time)
			ping.Error = "timeout"
			ping.Failed++
		} else {
			ping.Success++
			ping.Response = fmt.Sprintf("Reply from %v req=%v time=%v ms", ping.Target, ping.Success+ping.Failed, response.Time)

			ping.CTime = float64(response.Time)
			ping.TTime += ping.CTime
			ping.TTime2 += ping.CTime * ping.CTime
			if ping.TMin == 0 {
				ping.TMin = ping.CTime
			}
			if ping.CTime > ping.TMax {
				ping.TMax = ping.CTime
			}
			if ping.CTime < ping.TMin {
				ping.TMin = ping.CTime
			}

			if ping.Version == "" {
				ping.Version = response.Version
			}
			if ping.Version != response.Version {
				//not likely we'll see this happen but it doesnt hurt to be prepared
				fmt.Println("Host is sending back mismatched versions")
				fmt.Println("Old:", ping.Version, "New:", response.Version)
			}
		}
	} else {
		ping.Failed++
		ping.Error = response.Error
		ping.Response = response.Error
		return
	}
	return
}

func outputPing(Ping *Ping) {

	if Ping.Success > 0 {
		Ping.TAvg = Ping.TTime / Ping.Success
	}
	Ping.TTime2 /= Ping.Success

	if Ping.Success > 0 {
		Ping.TDev = math.Sqrt(Ping.TTime2 - Ping.TAvg*Ping.TAvg)
	}
	Ping.Percent = (Ping.Failed / Ping.Sent) * 100

	fmt.Println("\n---", Ping.Target, "ping statistics ---")
	fmt.Printf("%v packets transmitted, %v received, %.2f%% packet loss, time %vms\n", Ping.Sent, Ping.Success, Ping.Percent, Ping.TTime)
	fmt.Printf("rtt min/avg/max/mdev = %.3f/%.3f/%.3f/%.3f ms\n", Ping.TMin, Ping.TAvg, Ping.TMax, Ping.TDev)
	if Ping.Version != "" {
		fmt.Printf("Target is using cjdns version %v\n", Ping.Version)
	}
}
