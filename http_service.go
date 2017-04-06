// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

import (
	"log"
	"net/http"

	"github.com/christianparpart/sag/marathon"
)

// HttpService implements Service interface for HTTP services
type HttpService struct {
	ServiceId        string
	Scheduler        SchedulingAlgorithm
	Hosts            []string
	Backends         []*HttpBackend
	lastBackendIndex int
	selectBackend    func() *HttpBackend
}

func (service *HttpService) String() string {
	return service.ServiceId
}

func NewHttpService(serviceId string, Scheduler SchedulingAlgorithm, hosts []string) *HttpService {
	log.Printf("New service HTTP %v", serviceId)

	service := &HttpService{
		ServiceId: serviceId,
		Scheduler: Scheduler,
		Hosts:     hosts,
		Backends:  make([]*HttpBackend, 0),
	}

	service.selectBackend = service.RoundRobinScheduler

	return service
}

func (service *HttpService) Close() {
}

func (service *HttpService) IsEmpty() bool {
	return len(service.Backends) == 0
}

func (service *HttpService) AddBackend(id string, host string, port uint, capacity int, alive bool) {
	// XXX only add backend if not already present
	for _, backend := range service.Backends {
		if backend.Id == id {
			return
		}
	}

	backend := NewHttpBackend(id, host, port, capacity, alive)
	service.Backends = append(service.Backends, backend)
	log.Printf("New backend %v for %v with ID %v (%v)", backend, service.ServiceId, id, marathon.HealthStatus(alive))
}

func (service *HttpService) RemoveBackend(id string) {
	for i, backend := range service.Backends {
		if id == backend.Id {
			log.Printf("Remove backend %v from %v", backend, service)
			service.Backends = append(service.Backends[:i], service.Backends[i+1:]...)
			return
		}
	}
	log.Printf("No backend %v found in service %v", id, service)
}

func (service *HttpService) GetBackendById(id string) *HttpBackend {
	for _, backend := range service.Backends {
		if backend.Id == id {
			return backend
		}
	}
	return nil
}

func (service *HttpService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if backend := service.selectBackend(); backend != nil {
		backend.ServeHTTP(w, r)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func (service *HttpService) LeastLoadScheduler() *HttpBackend {
	i, leastLoaded := service.getFirstAvailableBackend()

	if leastLoaded != nil {
		for _, backend := range service.Backends[i:] {
			if backend.IsAvailable() && backend.CurrentLoad < leastLoaded.CurrentLoad {
				leastLoaded = backend
			}
		}
	}

	return leastLoaded
}

func (service *HttpService) RoundRobinScheduler() *HttpBackend {
	if len(service.Backends) == 0 {
		return nil
	}

	if service.lastBackendIndex+1 < len(service.Backends) {
		service.lastBackendIndex = service.lastBackendIndex + 1
	} else {
		service.lastBackendIndex = 0
	}

	return service.Backends[service.lastBackendIndex]
}

func (service *HttpService) ChanceScheduler() *HttpBackend {
	_, backend := service.getFirstAvailableBackend()
	return backend
}

func (service *HttpService) getFirstAvailableBackend() (int, *HttpBackend) {
	for i, backend := range service.Backends {
		if backend.IsAvailable() {
			return i, backend
		}
	}

	return len(service.Backends), nil
}
