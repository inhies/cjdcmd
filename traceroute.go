package main

import (
	"fmt"
	"github.com/inhies/go-cjdns/admin"
	"math"
	"sort"
)

type Routes []*Route

func (s Routes) Len() int      { return len(s) }
func (s Routes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByPath struct{ Routes }

func (s ByPath) Less(i, j int) bool { return s.Routes[i].RawPath < s.Routes[j].RawPath }

//Sorts with highest quality link at the top
type ByQuality struct{ Routes }

func (s ByQuality) Less(i, j int) bool { return s.Routes[i].RawLink > s.Routes[j].RawLink }

//TODO(inhies): Allow traceroute via path
func doTraceroute(user *admin.Admin, target string) {
	table := getTable(user)
	fmt.Println("Finding all routes to", target)

	count := 0
	for i := range table {

		if table[i].IP != target {
			continue
		}

		if table[i].Link < 1 {
			continue
		}

		response, err := getHops(table, table[i].RawPath)
		if err != nil {
			fmt.Println("Error:", err)
		}

		sort.Sort(ByPath{response})
		count++
		fmt.Printf("\nRoute #%d to target\n", count)
		for y, p := range response {

			fmt.Printf("IP: %v -- Version: %d -- Path: %s -- Link: %.0f -- Time:", p.IP, p.Version, p.Path, p.Link)
			if y == 0 {
				fmt.Printf(" Skipping ourself\n")
				continue
			}
			for x := 1; x <= 3; x++ {
				tRoute := &Ping{}
				tRoute.Target = p.Path
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
			println("")
		}
	}
	println("Found", count, "routes")
}

func getHops(table []*Route, fullPath uint64) (output []*Route, err error) {
	for i := range table {
		candPath := table[i].RawPath

		g := 64 - uint64(math.Log2(float64(candPath)))
		h := uint64(uint64(0xffffffffffffffff) >> g)

		if h&fullPath == h&candPath {
			output = append(output, table[i])
		}
	}
	return
}
