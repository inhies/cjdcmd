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
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/inhies/go-cjdns/admin"
	"github.com/spf13/cobra"
	"net"
	"os"
	"time"
)

func init() {
	TracerouteCmd.Flags().BoolVarP(&NmapOutput, "nmap", "x", false, "print result in nmap XML to stdout")
}

func tracerouteCmd(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		cmd.Usage()
		os.Exit(1)
	}

	var run *NmapRun
	startTime := time.Now()
	if NmapOutput {
		args := fmt.Sprint(os.Args[:])
		run = &NmapRun{
			Scanner:          "cjdmap",
			Args:             args,
			Start:            startTime.Unix(),
			Startstr:         startTime.String(),
			Version:          "0.1",
			XMLOutputVersion: "1.04",
			Hosts:            make([]*Host, 0, len(args)),
		}
	}

	targets := make([]*target, 0, len(args))
	for _, arg := range args {
		target, err := newTarget(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Skipping %s: %s\n", arg, err)
			continue
		}
		targets = append(targets, target)
	}

	c := Connect()
	table, err := c.NodeStore_dumpTable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to get routing table:", err)
	}
	table.SortByPath()

	for _, target := range targets {
		if NmapOutput {
			fmt.Fprintln(os.Stderr, target)
		} else {
			fmt.Fprintln(os.Stdout, target)
		}
		traces, err := target.trace(c, table)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to trace %s: %s\n", target, err)
			continue
		}
		if NmapOutput {
			run.Hosts = append(run.Hosts, traces[0])
		}
	}
	if NmapOutput {
		stopTime := time.Now()
		run.Finished = &Finished{
			Time:    stopTime.Unix(),
			TimeStr: stopTime.String(),
			//Elapsed: (stopTime.Sub(startTime) * time.Millisecond).String(),
			Exit: "success",
		}

		fmt.Fprint(os.Stdout, xml.Header)
		fmt.Fprintln(os.Stdout, `<?xml-stylesheet href="file:///usr/bin/../share/nmap/nmap.xsl" type="text/xsl"?>`)
		xEnc := xml.NewEncoder(os.Stdout)
		err = xEnc.Encode(run)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error:", err)
			os.Exit(1)
		}
	}
}

type target struct {
	addr net.IP
	name string
	rtt  time.Duration
	xml  *Host
}

func (t *target) String() string {
	if len(t.name) != 0 {
		return fmt.Sprintf("%s (%s)", t.name, t.addr)
	}
	return t.addr.String()
}

func newTarget(host string) (t *target, err error) {
	t = new(target)
	t.name, t.addr, err = resolve(host)
	return
}

var notInTableError = errors.New("not found in routing table")

func (t *target) trace(c *admin.Conn, table admin.Routes) (hostTraces []*Host, err error) {
	for _, r := range table {
		if t.addr.Equal(*r.IP) {
			hops := table.Hops(*r.Path)
			if hostTrace, err := t.traceHops(c, hops); err != nil {
				fmt.Fprintf(os.Stderr, "failed to trace %s, %s\n", r, err)
			} else {
				hostTraces = append(hostTraces, hostTrace)
				fmt.Fprintln(os.Stdout)
			}
		}
	}
	if len(hostTraces) == 0 {
		hostTraces = nil
		err = notInTableError
	}
	return
}

func (t *target) traceHops(c *admin.Conn, hops admin.Routes) (*Host, error) {
	hops.SortByPath()
	startTime := time.Now().Unix()
	trace := &Trace{Proto: "CJDNS"}

	for y, p := range hops {
		if y == 0 {
			continue
		}
		// Ping by path so we don't get RTT for a different route.
		rtt, _, err := c.RouterModule_pingNode(p.Path.String(), 0)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return nil, err
		}
		if rtt == 0 {
			rtt = 1
		}
		hostname, _, _ := resolve(p.IP.String())

		hop := &Hop{
			TTL:    y,
			RTT:    rtt,
			IPAddr: p.IP,
			Host:   hostname,
		}

		if NmapOutput {
			fmt.Fprintf(os.Stderr, "  %02d.% 4dms %s %s %s\n", y, rtt, p.Path, p.IP, hop.Host)
		} else {
			fmt.Fprintf(os.Stdout, "  %02d.% 4dms %s %s %s\n", y, rtt, p.Path, p.IP, hop.Host)
		}
		trace.Hops = append(trace.Hops, hop)
	}

	endTime := time.Now().Unix()
	h := &Host{
		StartTime: startTime,
		EndTime:   endTime,
		Status: &Status{
			State:     HostStateUp,
			Reason:    "pingNode",
			ReasonTTL: 56,
		},
		Address: &Address{Addr: &t.addr, AddrType: "ipv6"},
		Trace:   trace,
		//Times: &Times{ // Don't know what to do with this element yet.
		//	SRTT:   1,
		//	RTTVar: 1,
		//	To:     1,
		//},
	}

	if t.name != "" {
		h.Hostnames = []*Hostname{&Hostname{Name: t.name, Type: HostnameTypeUser}}
	}
	return h, nil
}
