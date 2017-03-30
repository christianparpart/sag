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
	"net/http"
	"net/http/httputil"
	"net/url"
)

type HttpBackend struct {
	id          string
	host        string
	port        uint
	alive       bool
	currentLoad int
	proxy       *httputil.ReverseProxy
}

func (backend *HttpBackend) String() string {
	return fmt.Sprintf("%v:%v", backend.host, backend.port)
}

func NewHttpBackend(id string, host string, port uint, alive bool) *HttpBackend {
	target, _ := url.Parse(fmt.Sprintf("http://%v:%v/", host, port))
	return &HttpBackend{
		id:    id,
		host:  host,
		port:  port,
		alive: alive,
		proxy: httputil.NewSingleHostReverseProxy(target),
	}
}

func (backend *HttpBackend) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// TODO: adjust alive stat, if failed (for up to N seconds, then reset)
	backend.proxy.ServeHTTP(rw, req)
}

func (backend *HttpBackend) SetAlive(alive bool) {
	if backend.alive != alive {
		backend.alive = alive

		if alive {
			log.Printf("Marking backend as alive. %v", backend)
		} else {
			log.Printf("Marking backend as dead. %v", backend)
		}
	}
}
