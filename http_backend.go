// This file is part of the "sad" project
//   <http://github.com/christianparpart/sad>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type HttpBackend struct {
	id          string
	host        string
	port        uint
	currentLoad int
	proxy       *httputil.ReverseProxy
}

func NewHttpBackend(id string, host string, port uint) *HttpBackend {
	target, _ := url.Parse(fmt.Sprintf("http://%v:%v/", host, port))
	return &HttpBackend{
		id:    id,
		host:  host,
		port:  port,
		proxy: httputil.NewSingleHostReverseProxy(target),
	}
}

func (backend *HttpBackend) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	backend.proxy.ServeHTTP(rw, req)
}
