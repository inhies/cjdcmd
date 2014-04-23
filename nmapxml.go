package main

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"net"
)

type HostState string

const (
	HostStateUp      = HostState("up")
	HostStateDown    = HostState("down")
	HostStateUnknown = HostState("unknown")
	HostStateSkipped = HostState("skipped")
)

type HostnameType string

const (
	HostnameTypeUser = HostnameType("user")
	HostnameTypePTR  = HostnameType("PTR")
)

type NmapRun struct {
	XMLName          xml.Name `xml:"nmaprun"`
	Scanner          string   `xml:"scanner,attr"`
	Args             string   `xml:"args,attr,omitempty"`
	Start            int64    `xml:"start,attr,omitempty"`
	Startstr         string   `xml:"startstr,attr,omitempty"`
	Version          string   `xml:"version,attr"`
	ProfileName      string   `xml:"porfile_name,attr,omitempty"`
	XMLOutputVersion string   `xml:"xmloutputversion,attr"`

	Hosts     []*Host   `xml:"host"`
	Finished  *Finished `xml:"runstats>finished"`
	HostStats *Hosts    `xml:"runstats>hosts"`
}

type Host struct {
	XMLName   xml.Name `xml:"host"`
	StartTime int64    `xml:"starttime,attr,omitempty"`
	EndTime   int64    `xml:"endtime,attr,omitempty"`
	Comment   string   `xml:"comment,attr,omitempty"`
	Status    *Status
	Address   *Address
	Hostnames []*Hostname `xml:"hostnames>hostname"`
	Trace     *Trace
	Times     *Times
}

type Status struct {
	XMLName   xml.Name  `xml:"status"`
	State     HostState `xml:"state,attr"`
	Reason    string    `xml:"reason,attr"`
	ReasonTTL int       `xml:"reason_ttl,attr"`
}

var UnknownHostStatus = &Status{State: HostStateUnknown}

type Address struct {
	XMLName  xml.Name `xml:"address"`
	Addr     *net.IP  `xml:"addr,attr"`
	AddrType string   `xml:"addrtype,attr"`
	Vendor   string   `xml:"vendor,attr,omitempty"`
}

type Hostname struct {
	Name string       `xml:"name,attr,omitempty"`
	Type HostnameType `xml:"type,attr,omitempty"`
}

type Trace struct {
	XMLName xml.Name `xml:"trace"`
	Proto   string   `xml:"proto,attr,omitempty"`
	Port    uint16   `xml:"port,attr,omitempty"`
	Hops    []*Hop
}

func (t *Trace) String() string {
	buf := new(bytes.Buffer)
	for i, hop := range t.Hops {
		fmt.Fprintf(buf, "  %02d. %dms %s %s\n", i, hop.RTT, hop.IPAddr, hop.Host)
	}
	return buf.String()
}

type Hop struct {
	XMLName xml.Name `xml:"hop"`
	TTL     int      `xml:"ttl,attr"`
	RTT     int      `xml:"rtt,attr,omitempty"`
	IPAddr  *net.IP  `xml:"ipaddr,attr,omitempty"`
	Host    string   `xml:"host,attr,omitempty"`
}

type Times struct {
	XMLName xml.Name `xml:"times"`
	SRTT    int64    `xml:"srtt,attr"`
	RTTVar  int64    `xml:"rttvar,attr"`
	To      int64    `xml:"to,attr"`
}

type Finished struct {
	XMLName  xml.Name `xml:"finished"`
	Time     int64    `xml:"time,attr"`
	TimeStr  string   `xml:"timestr,attr,omitempty"`
	Elapsed  float64  `xml:"elapsed,attr"`
	Summary  string   `xml:"summary,attr,omitempty"`
	Exit     string   `xml:"exit,attr,omitempty"`
	ErrorMsg string   `xml:"errormsg,attr,omitempty"`
}

type Hosts struct {
	Up    int `xml:"up,attr"`
	Down  int `xml:"down,attr"`
	Total int `xml:"total,attr"`
}
