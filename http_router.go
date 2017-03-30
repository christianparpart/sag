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
)

type HttpRouter struct {
	ListenAddr string
	GetService func(*http.Request) *HttpService
}

func (router HttpRouter) Run() {
	err := http.ListenAndServe(router.ListenAddr, router)
	if err != nil {
		log.Fatal(err)
	}
}

func (router HttpRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if service := router.GetService(r); service != nil {
		service.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "No service found for request host header %q\n", r.Host)
	}
}
