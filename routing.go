// Package routing provides utilities to read and parse the Linux routing table.
// It allows retrieving the default gateway and associated network interface by
// reading data from /proc/net/route and interpreting route flags.

package routing

import (
	"errors"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

// RoutingTable represents a single entry in the Linux routing table.
// It contains details about network routes, including the interface, destination, and gateway.
type RoutingTable struct {
	Interface   string      // The network interface associated with the route.
	Destination string      // The destination IP address for the route.
	Gateway     string      // The gateway IP address for the route.
	Flags       []RouteFlag // Flags associated with the route.
	RefCnt      int8        // Reference count for the route.
	Use         int8        // Usage count of the route.
	Metric      int8        // Metric for the route, used in route selection.
	Mask        string      // The subnet mask for the route.
	MTU         int8        // Maximum transmission unit for the route.
	Window      int8        // Window size for the route.
	IRTT        int8        // Initial round trip time for the route.
}

// RouteFlag represents a flag used in routing, indicating specific route characteristics.
type RouteFlag struct {
	Letter string // Symbol representing the flag (e.g., "U" for up, "G" for gateway).
	Bit    int16  // Bitmask value for the flag.
	Name   string // Full name of the flag.
	Desc   string // Description of what the flag indicates.
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

// DecimalToIP converts a decimal integer into its equivalent IPv4 address format.
// It takes a decimal integer and converts it to a human-readable IP address string.
func DecimalToIP(decimal int64) string {
	ip := net.IPv4(
		byte(decimal),
		byte(decimal>>8),
		byte(decimal>>16),
		byte(decimal>>24),
	)

	return ip.String() // Returns the IP address as a string.
}

// computeRouteFlag takes a bitmask and generates a list of RouteFlags based on it.
// It takes a bitmask as input and returns the corresponding RouteFlags.
func computeRouteFlag(bits int16) []RouteFlag {
	rf := make([]RouteFlag, 0)
	var counter int16 = 1

	for i := range make([]int16, bits) {
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

	return rf // Returns the list of RouteFlags corresponding to the bitmask.
}

// GetLinuxRoutingTable retrieves the current routing table from the Linux operating system.
// It reads the routing information from /proc/net/route and populates a slice of RoutingTable structs.
func GetLinuxRoutingTable(table *[]RoutingTable) error {
	f, fErr := os.Open("/proc/net/route")
	if fErr != nil {
		return errors.New(fErr.Error()) // Returns an error if the file cannot be opened.
	}

	b, bErr := io.ReadAll(f)
	if bErr != nil {
		return errors.New(bErr.Error()) // Returns an error if reading the file fails.
	}

	defer f.Close() // Ensures the file is closed when the function exits.

	fTable := string(b)
	fRows := strings.Split(fTable, "\n")         // Splits the file content into rows.
	description := strings.Split(fRows[0], "\t") // Gets the header for routing table entries.

	for _, v := range fRows {
		if strings.Contains(v, "Iface") {
			continue // Skip the header row.
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
					return errors.New(valErr.Error()) // Returns an error if converting the gateway address fails.
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
		*table = append(*table, rtRow) // Append the populated RoutingTable struct to the slice.
	}

	return nil // Return nil if the operation completes successfully.
}

// flagContains checks if a slice of RouteFlags contains a specific flag letter.
// It returns true if the flag is found, otherwise false.
func flagContains(rf []RouteFlag, letter string) bool {
	for _, v := range rf {
		if strings.Contains(v.Letter, letter) {
			return true // Flag letter found.
		}
	}
	return false // Flag letter not found.
}

// getDefaultGW returns the RoutingTable entry that contains the default gateway.
// It searches the routing table for an entry marked with the "U" (up) and "G" (gateway) flags.
func getDefaultGW() (RoutingTable, error) {
	rt := new([]RoutingTable)

	err := GetLinuxRoutingTable(rt)
	if err != nil {
		if len(*rt) > 0 {
			return (*rt)[0], nil // Return the first entry if error occurs but entries are present.
		}
		return RoutingTable{}, errors.New(err.Error()) // Return error if no entries are present.
	}

	up := false
	gateway := false
	for _, v := range *rt {
		if flagContains(v.Flags, "U") {
			up = true
		}
		if flagContains(v.Flags, "G") {
			gateway = true
		}

		if up && gateway {
			return v, nil // Return the entry with both "U" and "G" flags.
		}
	}
	return RoutingTable{}, errors.New("could not locate default GW") // Error if default GW not found.
}

// FindLinuxDefaultGW retrieves the default gateway address by reading the routing table.
// It returns the default gateway IP address in standard string format.
func FindLinuxDefaultGW() (string, error) {
	tr, err := getDefaultGW()
	if err != nil {
		return "", errors.New(err.Error()) // Return error if default GW not found.
	}

	return tr.Gateway, nil // Return the default gateway IP address.
}

// FindLinuxDefaultGWInterface returns the network interface name of the default gateway.
// It reads the routing table to find the interface associated with the default gateway.
func FindLinuxDefaultGWInterface() (string, error) {
	tr, err := getDefaultGW()
	if err != nil {
		return "", errors.New(err.Error()) // Return error if default GW interface not found.
	}

	return tr.Interface, nil // Return the network interface name of the default gateway.
}
