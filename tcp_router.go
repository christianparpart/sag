package main

import (
	"log"
	"net"
)

type TcpRouter struct {
	ListenAddr string
	listener   net.Listener
	getService func(net.Conn) *TcpService
}

func NewTcpRouter(laddr string) (*TcpRouter, error) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return nil, err
	}
	router := &TcpRouter{
		ListenAddr: laddr,
		listener:   listener,
	}
	return router, nil
}

func (router *TcpRouter) Close() {
	router.listener.Close()
}

func (router *TcpRouter) Serve() {
	for {
		conn, err := router.listener.Accept()
		if err != nil {
			log.Printf("Failed to accept TCP listener %v. %v", router.ListenAddr, err)
			continue
		}
		service := router.getService(conn)
		if service != nil {
			log.Printf("Router %v failed to route TCP connection %v to service.",
				router.ListenAddr, conn.RemoteAddr())
			conn.Close()
			continue
		}
		go service.ServeTCP(conn)
	}
}
