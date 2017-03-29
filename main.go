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
	"os"
	"time"
)

type ServiceApplicationGateway struct {
	Discoveries  []Discovery
	EventStream  chan interface{}
	HttpServices map[string]*HttpService
}

func (sag *ServiceApplicationGateway) FindHttpServiceById(serviceId string) *HttpService {
	if service, ok := sag.HttpServices[serviceId]; ok {
		return service
	}
	return nil
}

func (sag *ServiceApplicationGateway) FindHttpServiceByHost(host string) *HttpService {
	for _, s := range sag.HttpServices {
		for _, shost := range s.Hosts {
			if host == shost {
				return s
			}
		}
	}

	return nil
}

func (sag *ServiceApplicationGateway) Register(sd Discovery) {
	sag.Discoveries = append(sag.Discoveries, sd)
	go sd.Run()
}

func (sag *ServiceApplicationGateway) HandleEvents() {
	for {
		switch v := (<-sag.EventStream).(type) {
		case AddHttpServiceEvent:
			if _, ok := sag.HttpServices[v.ServiceId]; !ok {
				sag.HttpServices[v.ServiceId] = &HttpService{
					ServiceId: v.ServiceId,
					Hosts:     v.Hosts,
				}
			}
		case AddBackendEvent:
			if service, ok := sag.HttpServices[v.ServiceId]; ok {
				service.AddBackend(v.BackendId, v.Hostname, v.Port, v.Alive)
			}
		case HealthStatusChangedEvent:
		case RemoveBackendEvent:
			if service, ok := sag.HttpServices[v.ServiceId]; ok {
				service.RemoveBackend(v.BackendId)
				if service.IsEmpty() {
					delete(sag.HttpServices, v.ServiceId)
				}
			} else {
				log.Printf("RemoveBackendEvent: %v %v", v.ServiceId, v.BackendId)
			}
		case LogEvent:
			log.Print(v.Message)
		}
	}
}

func (sag *ServiceApplicationGateway) RunHttpGateway(addr net.IP, port int) {
	httpGateway := HttpGateway{
		ListenAddr: fmt.Sprintf("%v:%v", addr, port),
		GetService: sag.getService,
	}
	httpGateway.Run()
}

func (sag *ServiceApplicationGateway) getService(r *http.Request) *HttpService {
	return sag.FindHttpServiceByHost(r.Host)
}

func (sag *ServiceApplicationGateway) Shutdown() {
	// TODO: gracefully shutdown
	// stop accepting new sessions, gracefully terminate active sessions
}

func main() {
	sag := ServiceApplicationGateway{
		EventStream:  make(chan interface{}),
		HttpServices: make(map[string]*HttpService),
	}

	// add a service discovery source
	port := uint(Atoi(os.Getenv("MARATHON_PORT"), 8080))
	host := os.Getenv("MARATHON_IP")
	if len(host) == 0 {
		host = "127.0.0.1"
	}
	sag.Register(NewDiscoveryMarathon(net.ParseIP(host), port, time.Second*1, sag.EventStream))

	// handle any incoming service discovery events
	go sag.HandleEvents()

	sag.RunHttpGateway(net.ParseIP("0.0.0.0"), 8080)

	sag.Shutdown()
}
