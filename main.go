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
	"time"

	flag "github.com/ogier/pflag"
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

func (sag *ServiceApplicationGateway) getHttpServiceByHost(r *http.Request) *HttpService {
	for _, s := range sag.HttpServices {
		for _, shost := range s.Hosts {
			if r.Host == shost {
				return s
			}
		}
	}

	return nil
}

func (sag *ServiceApplicationGateway) RegisterDiscovery(sd Discovery) {
	sag.Discoveries = append(sag.Discoveries, sd)
	go sd.Run()
}

func (sag *ServiceApplicationGateway) ProcessEvents() {
	for {
		switch v := (<-sag.EventStream).(type) {
		case RestoreFromSnapshotEvent:
			log.Printf("Start restoring state from snapshot")
		case AddHttpServiceEvent:
			if _, ok := sag.HttpServices[v.ServiceId]; !ok {
				sag.HttpServices[v.ServiceId] = NewHttpService(v.ServiceId, v.Hosts)
			}
		case AddBackendEvent:
			if service, ok := sag.HttpServices[v.ServiceId]; ok {
				service.AddBackend(v.BackendId, v.Hostname, v.Port, v.Alive)
			}
		case HealthStatusChangedEvent:
			if service := sag.FindHttpServiceById(v.ServiceId); service != nil {
				if backend := service.GetBackendById(v.BackendId); backend != nil {
					backend.SetAlive(v.Alive)
				}
			}
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

func (sag *ServiceApplicationGateway) RunHttpRouterByHost(addr net.IP, port uint) {
	HttpRouter{
		ListenAddr: fmt.Sprintf("%v:%v", addr, port),
		GetService: sag.getHttpServiceByHost,
	}.Run()
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

	httpVhostIP := flag.IP("http-vhost-ip", net.ParseIP("0.0.0.0"), "HTTP vhost router bind IP")
	httpVhostPort := flag.Uint("http-vhost-port", 8080, "HTTP vhost router port number")
	marathonIP := flag.IP("marathon-ip", net.ParseIP("127.0.0.1"), "Marathon IP address")
	marathonPort := flag.Uint("marathon-port", 8080, "Marathon port number")
	flag.Parse()

	// add a service discovery source
	sag.RegisterDiscovery(NewDiscoveryMarathon(*marathonIP, *marathonPort, time.Second*1, sag.EventStream))

	// add router (HTTP application by-vhost router)
	go sag.RunHttpRouterByHost(*httpVhostIP, *httpVhostPort)

	// process any incoming service discovery events
	sag.ProcessEvents()

	sag.Shutdown()
}
