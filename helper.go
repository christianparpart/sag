// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

import (
	"net"
	"strconv"
	"strings"
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

func makeStringArray(s string) []string {
	if len(s) == 0 {
		return []string{}
	} else {
		return strings.Split(s, ",")
	}
}

func MakeBool(s string) bool {
	switch s {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}
