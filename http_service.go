// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

import (
	"net/http"
)

// HttpService implements Service interface for HTTP services
type HttpService struct {
	ServiceId        string
	Hosts            []string
	backends         []*HttpBackend
	lastBackendIndex int
}

func NewHttpService(serviceId string, hosts []string) *HttpService {
	return &HttpService{
		ServiceId: serviceId,
		Hosts:     hosts,
	}
}

func (s *HttpService) IsEmpty() bool {
	return len(s.backends) == 0
}

func (s *HttpService) AddBackend(id string, host string, port uint, alive bool) {
	// XXX only add backend if not already present
	for _, backend := range s.backends {
		if backend.id == id {
			return
		}
	}

	s.backends = append(s.backends, NewHttpBackend(id, host, port, alive))
}

func (s *HttpService) RemoveBackend(id string) {
	for i, backend := range s.backends {
		if id == backend.id {
			s.backends = append(s.backends[:i], s.backends[i+1:]...)
			return
		}
	}
}

func (s *HttpService) SelectBackend() *HttpBackend {
	if len(s.backends) == 0 {
		return nil
	}

	// XXX do simple round-robin for now
	if s.lastBackendIndex+1 < len(s.backends) {
		s.lastBackendIndex = s.lastBackendIndex + 1
	} else {
		s.lastBackendIndex = 0
	}

	return s.backends[s.lastBackendIndex]
}

func (s *HttpService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if backend := s.SelectBackend(); backend != nil {
		backend.ServeHTTP(w, r)
	}
}
