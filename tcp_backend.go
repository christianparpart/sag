package main

import (
	"net"
)

type TcpBackend struct {
	Id          string
	Host        string
	Port        uint
	Capacity    int
	CurrentLoad int
	Alive       bool
	proxy       *TcpProxy
}

func NewTcpBackend(id string, ip net.IP, port uint, capacity int) *TcpBackend {
	return &TcpBackend{
		Id:          id,
		Host:        ip.String(),
		Port:        port,
		Capacity:    capacity,
		CurrentLoad: 0,
		proxy:       NewTcpProxy(ip, port),
	}
}

func (backend *TcpBackend) ServeTCP(conn net.Conn) {
	backend.CurrentLoad++

	conn.Close()

	backend.CurrentLoad--
}
