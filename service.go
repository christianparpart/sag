// This file is part of the "sad" project
//   <http://github.com/christianparpart/sad>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

// Service is an interface, generic enough to cover any kind of network service,
// providing the ability to add and remove backends.
type Service interface {
	AddBackend(name string, host string, port int)
	RemoveBackend(name string)
}

type Backend interface {
	Name() string
	Host() string
	Port() int
}
