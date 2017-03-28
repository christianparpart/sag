package main

import (
	"log"
	"net"
	"os"
	"time"
)

type ServiceApplicationGateway struct {
	ServiceDiscoveries []ServiceDiscovery
	EventStream        chan interface{}
}

func (sag *ServiceApplicationGateway) Register(sd ServiceDiscovery) {
	sag.ServiceDiscoveries = append(sag.ServiceDiscoveries, sd)
	go sd.Run()
}

func (sag *ServiceApplicationGateway) HandleEvents() {
	for {
		event := <-sag.EventStream
		switch v := event.(type) {
		case AddBackendEvent:
			log.Printf("AddBackendEvent: %v %v %v:%v", v.ServiceId, v.BackendId, v.Hostname, v.Port)
		case RemoveBackendEvent:
			log.Printf("RemoveBackendEvent: %v %v", v.ServiceId, v.BackendId)
		case LogEvent:
			log.Print(v.Message)
		}
	}
}

func (sag *ServiceApplicationGateway) Shutdown() {
	// TODO: gracefully shutdown
	// stop accepting new sessions, gracefully terminate active sessions
}

func main() {
	sag := ServiceApplicationGateway{
		EventStream: make(chan interface{}),
	}

	// add a service discovery source
	host := os.Getenv("MARATHON_IP")
	port := Atoi(os.Getenv("MARATHON_PORT"), 8080)
	sag.Register(NewMarathonSD(net.ParseIP(host), port, time.Second*1, sag.EventStream))

	// TODO sag.AddHttpGateway(net.ParseIP("0.0.0.0"), 80)

	// handle any incoming service discovery events
	sag.HandleEvents()

	sag.Shutdown()
}
