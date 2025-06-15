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
	Flags       int8
	RefCnt      int8
	Use         int8
	Metric      int8
	Mask        string
	MTU         int8
	Window      int8
	IRTT        int8
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

// In Linux there is the /proc/net/route file, it contains
// all the routing information defined on the Linux OS,
// this Function returns the default gateway address by
// reading the file and converting the HEX to DEC and DEC to IP
func FindLinuxDefaultGW() (string, error) {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return "", errors.New(err.Error())
	}

	b, errRead := io.ReadAll(f)
	if errRead != nil {
		return "", errors.New(errRead.Error())
	}

	defer f.Close()

	table := string(b)
	rows := strings.Split(table, "\n")

	description := strings.Split(rows[0], "\t")

	var gw_at int

	for i, v := range description {
		if strings.TrimSpace(v) == "Gateway" {
			gw_at = i
			break
		}
	}

	var decimal int64
	var valErr error

	for row := range rows {
		if !strings.Contains(strings.TrimSpace(rows[row]), "Iface") {
			row_vals := strings.Split(rows[row], "\t")
			gw_val := strings.TrimSpace(row_vals[gw_at])
			// If the Gateway is not set, iterate to next row
			if gw_val != "00000000" {
				decimal, valErr = strconv.ParseInt(gw_val, 16, 64)
				if valErr != nil {
					return "", errors.New(valErr.Error())
				}
				break
			}
		}
	}

	return DecimalToIP(decimal), nil
}
