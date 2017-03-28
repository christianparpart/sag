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
