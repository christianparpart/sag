package main

import (
	"net"
	"strconv"
)

var resolveMap = make(map[string]string)

func SoftResolveIPAddr(dns string) string {
	if value, ok := resolveMap[dns]; ok {
		return value
	}

	if ip, err := net.ResolveIPAddr("ip", dns); err == nil {
		return ip.String()
	} else {
		// fallback to actual dns name
		return dns
	}
}

func Atoi(value string, defaultValue int) int {
	if result, err := strconv.Atoi(value); err == nil {
		return result
	}

	return defaultValue
}
