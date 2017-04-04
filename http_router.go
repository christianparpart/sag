// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
)

type HttpRouter struct {
	ListenAddr string
	listener   net.Listener
	getService func(*http.Request) *HttpService
}

func NewHttpRouter(laddr string, getService func(*http.Request) *HttpService) *HttpRouter {
	return &HttpRouter{
		ListenAddr: laddr,
		listener:   nil,
		getService: getService,
	}
}

func (router *HttpRouter) Run() {
	listener, err := net.Listen("tcp", router.ListenAddr)
	if err != nil {
		log.Fatal(err)
	}

	router.listener = listener

	err = http.Serve(listener, router)
	if err != nil {
		log.Fatal(err)
	}
}

func (router *HttpRouter) Close() {
	router.listener.Close()
}

func (router HttpRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if service := router.getService(r); service != nil {
		service.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "No service found for request host header %q\n", r.Host)
	}
}
