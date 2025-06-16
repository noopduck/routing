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
				flag, _ = strconv.ParseInt(v, 10, 8)
				rtRow.Flags = int8(flag)
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
