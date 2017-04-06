package main

import (
	"net"
)

type TcpService struct {
	ServiceId        string
	Scheduler        SchedulingAlgorithm
	Backends         []*TcpBackend
	lastBackendIndex int
	selectBackend    func() *TcpBackend
}

func (service *TcpService) ServeTCP(conn net.Conn) {
	// TODO: selectBackend()
	defer conn.Close()
}
