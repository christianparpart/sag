package main

import (
	"fmt"
	"log"
	"net"
)

type TcpProxy struct {
	IP   net.IP
	Port uint
}

func NewTcpProxy(ip net.IP, port uint) *TcpProxy {
	return &TcpProxy{
		IP:   ip,
		Port: port,
	}
}

func (proxy *TcpProxy) ServeTCP(conn net.Conn) {
	log.Printf("TODO: TcpProxy(%v).ServeTCP()", proxy)
	conn.Close()
}

func (proxy *TcpProxy) String() string {
	return fmt.Sprintf("%v:%v", proxy.IP, proxy.Port)
}
