// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

import "fmt"

type Protocol int

const (
	HTTP Protocol = iota
	TCP
	UDP
)

func (p Protocol) String() string {
	switch p {
	case HTTP:
		return "HTTP"
	case TCP:
		return "TCP"
	case UDP:
		return "UDP"
	default:
		return fmt.Sprintf("<%v>", int(p))
	}
}

// Service is an interface, generic enough to cover any kind of network service,
// providing the ability to add and remove backends.
type Service interface {
	AddBackend(id string, host string, port uint, capacity int, alive bool)
	RemoveBackend(id string)
}

type Backend interface {
	Id() string

	Host() string
	Port() uint

	GetCurrentLoad() int
	GetCapacity() int

	Alive() bool
}
