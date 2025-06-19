package routing

import (
	"errors"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// Struct representing the content inside
// /proc/net/route
type RoutingTable struct {
	Interface   string
	Destination string
	Gateway     string
	Flags       []RouteFlag
	RefCnt      int8
	Use         int8
	Metric      int8
	Mask        string
	MTU         int8
	Window      int8
	IRTT        int8
}

// RouteFlag represents a routing flag and its meaning.
type RouteFlag struct {
	Letter string // Symbol, e.g., "U", "G"
	Bit    int16  // Bitmask value
	Name   string // Full name
	Desc   string // Description of what the flag means
}

var routeFlags = []RouteFlag{
	{"U", 0x1, "Up", "Route is usable (interface is up)"},
	{"G", 0x2, "Gateway", "Destination is a gateway"},
	{"H", 0x4, "Host", "Target is a host (not a network)"},
	{"R", 0x8, "Reinstate", "Route was reinstated for dynamic routing"},
	{"D", 0x10, "Dynamic", "Route was dynamically created by daemon or redirect"},
	{"M", 0x20, "Modified", "Route was modified by redirect"},
	{"A", 0x40, "Addrconf", "Route created by address autoconf"},
	{"C", 0x80, "Cache", "Route is in cache"},
}

// Function to return the IP address
// on it's normal form
func DecimalToIP(decimal int64) string {
	ip := net.IPv4(
		byte(decimal),
		byte(decimal>>8),
		byte(decimal>>16),
		byte(decimal>>24),
	)

	return ip.String()
}

// Assign the respective flags in a more defined way
func computeRouteFlag(bits int16) []RouteFlag {
	rf := make([]RouteFlag, 0)
	var counter int16 = 1

	for i := int16(0); i < bits; i++ {
		if counter == routeFlags[i].Bit {
			rf = append(rf, RouteFlag{
				Letter: routeFlags[i].Letter,
				Bit:    routeFlags[i].Bit,
				Name:   routeFlags[i].Name,
				Desc:   routeFlags[i].Desc,
			})
			counter++
		}
	}

	return rf
}

// This returns the complete routing table from the LinuxOS
func GetLinuxRoutingTable() ([]RoutingTable, error) {
	f, fErr := os.Open("/proc/net/route")
	if fErr != nil {
		return nil, errors.New(fErr.Error())
	}

	b, bErr := io.ReadAll(f)
	if bErr != nil {
		return nil, errors.New(bErr.Error())
	}

	defer f.Close()

	fTable := string(b)
	fRows := strings.Split(fTable, "\n")
	description := strings.Split(fRows[0], "\t")

	table := make([]RoutingTable, 0)

	for _, v := range fRows {
		if strings.Contains(v, "Iface") {
			continue
		}
		fColumn := strings.Split(v, "\t")
		rtRow := RoutingTable{}
		for n, v := range fColumn {
			d := strings.TrimSpace(description[n])
			switch d {
			case "Iface":
				rtRow.Interface = v
			case "Destination":
				rtRow.Destination = v
			case "Gateway":
				val, valErr := strconv.ParseInt(v, 16, 64)
				if valErr != nil {
					return nil, errors.New(valErr.Error())
				}
				rtRow.Gateway = DecimalToIP(val)
			case "Flags":
				var flag int64
				flag, _ = strconv.ParseInt(v, 10, 16)
				rtRow.Flags = computeRouteFlag(int16(flag))
			case "RefCnt":
				var refcnt int64
				refcnt, _ = strconv.ParseInt(v, 10, 8)
				rtRow.RefCnt = int8(refcnt)
			case "Use":
				var use int64
				use, _ = strconv.ParseInt(v, 10, 8)
				rtRow.Use = int8(use)
			case "Metric":
				var metric int64
				metric, _ = strconv.ParseInt(v, 10, 8)
				rtRow.Metric = int8(metric)
			case "Mask":
				rtRow.Mask = v
			case "MTU":
				var mtu int64
				mtu, _ = strconv.ParseInt(v, 10, 8)
				rtRow.MTU = int8(mtu)
			case "Window":
				var window int64
				window, _ = strconv.ParseInt(v, 10, 8)
				rtRow.Window = int8(window)
			case "IRTT":
				var irtt int64
				irtt, _ = strconv.ParseInt(v, 10, 8)
				rtRow.IRTT = int8(irtt)
			}
		}
		table = append(table, rtRow)
	}

	//	var table []RoutingTable = make([]RoutingTable, 0)
	return table, nil
}

func flagContains(rf []RouteFlag, letter string) bool {
	for _, v := range rf {
		if strings.Contains(v.Letter, letter) {
			return true
		}
	}
	return false
}

func getDefaultGW() (RoutingTable, error) {
	rt, err := GetLinuxRoutingTable()
	if err != nil {
		return rt[0], errors.New(err.Error())
	}

	up := false
	gateway := false
	for _, v := range rt {
		if flagContains(v.Flags, "U") {
			up = true
		}
		if flagContains(v.Flags, "G") {
			gateway = true
		}

		if up && gateway {
			return v, nil
		}
	}
	return rt[0], errors.New("could not locate default GW")
}

// In Linux there is the /proc/net/route file, it contains
// all the routing information defined on the Linux OS,
// this Function returns the default gateway address by
// reading the file and converting the HEX to DEC and DEC to IP
func FindLinuxDefaultGW() (string, error) {
	tr, err := getDefaultGW()
	if err != nil {
		return "", errors.New(err.Error())
	}

	return tr.Gateway, nil
}

func FindLinuxDefaultGGWInterface() (string, error) {
	tr, err := getDefaultGW()
	if err != nil {
		return "", errors.New(err.Error())
	}

	return tr.Interface, nil
}
