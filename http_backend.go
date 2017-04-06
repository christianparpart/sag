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
)

type HttpBackend struct {
	Id          string
	Host        string
	Port        uint
	Capacity    int
	CurrentLoad int
	Alive       bool
	ServedTotal uint64
	proxy       *httputil.ReverseProxy
}

func NewHttpBackend(id string, host string, port uint, capacity int, alive bool) *HttpBackend {
	targetHost := fmt.Sprintf("%v:%v", host, port)
	director := func(req *http.Request) {
		req.URL.Scheme = "http"
		req.URL.Host = targetHost
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "") // explicitely disable (avoid defaulting)
		}
	}
	modifyResponse := func(rw *http.Response) error {
		via := fmt.Sprintf("%v.%v sag", rw.Request.ProtoMajor, rw.Request.ProtoMinor)
		rw.Header.Add("Via", via)
		return nil
	}
	proxy := &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modifyResponse,
	}

	backend := &HttpBackend{
		Id:          id,
		Host:        host,
		Port:        port,
		Capacity:    capacity,
		CurrentLoad: 0,
		Alive:       alive,
		proxy:       proxy,
	}

	return backend
}

func (backend *HttpBackend) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// TODO: adjust alive stat, if failed (for up to N seconds, then reset)

	backend.ServedTotal++

	backend.CurrentLoad++
	backend.proxy.ServeHTTP(rw, req)
	backend.CurrentLoad--
}

func (backend *HttpBackend) IsAvailable() bool {
	return backend.Alive && (backend.Capacity == 0 || backend.CurrentLoad < backend.Capacity)
}

func (backend *HttpBackend) SetAlive(alive bool) {
	if backend.Alive != alive {
		backend.Alive = alive

		if alive {
			log.Printf("Backend is alive. %v", backend)
		} else {
			log.Printf("Backend is dead. %v", backend)
		}
	}
}

func (backend *HttpBackend) String() string {
	return fmt.Sprintf("%v:%v", backend.Host, backend.Port)
}
