package main

import (
	"fmt"
	"github.com/inhies/go-cjdns/admin"
	"math"
)

type Ping struct {
	Target, Version, Response, Error             string
	Failed, Percent, Sent, Success               float64
	CTime, TTime, TTime2, TMin, TAvg, TMax, TDev float64
}

// Pings a node and generates statistics
func pingNode(user *admin.Admin, ping *Ping) (err error) {
	response, err := admin.RouterModule_pingNode(user, ping.Target, PingTimeout)

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
			ping.Response = fmt.Sprintf("Reply from %v %vms", ping.Target, response.Time)
			ping.Success++
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
				println("Host is sending back mismatched versions")
			}
		}
	} else {
		ping.Failed++
		err = fmt.Errorf(response.Error)
		ping.Error = response.Error
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
	fmt.Printf("Target is using cjdns version %v\n", Ping.Version)
}
