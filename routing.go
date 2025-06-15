package routing

import (
	"errors"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
)

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

	gw_val := strings.Split(rows[1], "\t")

	decimal, valErr := strconv.ParseInt(gw_val[gw_at], 16, 64)
	if valErr != nil {
		return "", errors.New(valErr.Error())
	}

	return DecimalToIP(decimal), nil
}
