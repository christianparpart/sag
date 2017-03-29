// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/christianparpart/sag/marathon"
)

const (
	LB_PROXY_PROTOCOL      = "lb-proxy-protocol"
	LB_ACCEPT_PROXY        = "lb-accept-proxy"
	LB_VHOST_HTTP          = "lb-vhost"
	LB_VHOST_DEFAULT_HTTP  = "lb-vhost-default"
	LB_VHOST_HTTPS         = "lb-vhost-ssl"
	LB_VHOST_DEFAULT_HTTPS = "lb-vhost-default-ssl"
)

type Discovery interface {
	Run()
	Shutdown()
}

type DiscoveryMarathon struct {
	marathonIP   net.IP
	marathonPort uint
	sse          *EventSource
	eventStream  chan<- interface{}
}

func NewDiscoveryMarathon(host net.IP, port uint, reconnectDelay time.Duration, eventStream chan<- interface{}) *DiscoveryMarathon {
	url := fmt.Sprintf("http://%v:%v/v2/events", host, port)
	sse := NewEventSource(url, reconnectDelay)

	sd := &DiscoveryMarathon{
		marathonIP:   host,
		marathonPort: port,
		sse:          sse,
		eventStream:  eventStream}

	sse.OnOpen = sd.onOpen
	sse.OnError = sd.onError
	sse.AddEventListener("status_update_event", sd.status_update_event)
	sse.AddEventListener("health_status_changed_event", sd.health_status_changed_event)
	sse.OnMessage = func(event, data string) {
		sd.log(fmt.Sprintf("SSE[%v]: %v", event, data))
	}

	return sd
}

func (sd *DiscoveryMarathon) String() string {
	return fmt.Sprintf("DiscoveryMarathon<%v>", sd.sse.Url)
}

func (sd *DiscoveryMarathon) Run() {
	sd.log("Starting")
	sd.sse.RunForever()
}

func (sd *DiscoveryMarathon) Shutdown() {
	sd.sse.Close()
}

func (sd *DiscoveryMarathon) log(msg string) {
	sd.eventStream <- LogEvent{
		Message: fmt.Sprintf("marathon(%v): %v", sd.sse.Url, msg)}
}

func (sd *DiscoveryMarathon) onOpen() {
	sd.log("Connected")
	sd.LoadAllApps()
}

func (sd *DiscoveryMarathon) onError(message string) {
	sd.log("SSE failure. " + message)
}

func (sd *DiscoveryMarathon) status_update_event(data string) {
	var event marathon.StatusUpdateEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		sd.log(fmt.Sprintf("Failed to unmarshal status_update_event. %v", err))
		sd.log(fmt.Sprintf("status_update_event: %+v\n", data))
	} else {
		sd.statusUpdateEvent(&event)
	}
}

func (sd *DiscoveryMarathon) statusUpdateEvent(event *marathon.StatusUpdateEvent) {
	switch event.TaskStatus {
	case marathon.TaskRunning:
		// add task iff no health checks are available
		app, _ := sd.getMarathonApp(event.AppId)
		if app != nil && len(app.HealthChecks) == 0 {
			for portIndex, portDef := range app.PortDefinitions {
				serviceId := makeServiceId(event.AppId, portIndex)
				sd.eventStream <- AddHttpServiceEvent{
					ServiceId: serviceId,
					Hosts:     makeStringArray(portDef.Labels[LB_VHOST_HTTP]),
				}
				sd.eventStream <- AddBackendEvent{
					ServiceId: serviceId,
					BackendId: event.TaskId,
					Hostname:  event.Host,
					Port:      event.Ports[portIndex],
				}
			}
		}
	case marathon.TaskFinished, marathon.TaskFailed, marathon.TaskKilling, marathon.TaskKilled, marathon.TaskLost:
		for portIndex, _ := range event.Ports {
			sd.eventStream <- RemoveBackendEvent{
				ServiceId: makeServiceId(event.AppId, portIndex),
				BackendId: event.TaskId,
			}
		}
	}
}

func (sd *DiscoveryMarathon) health_status_changed_event(data string) {
	var event marathon.HealthStatusChangedEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		sd.log(fmt.Sprintf("Failed to unmarshal health_status_changed_event. %v\n", err))
	} else {
		sd.healthStatusChangedEvent(&event)
	}
}

func (sd *DiscoveryMarathon) healthStatusChangedEvent(event *marathon.HealthStatusChangedEvent) {
	// if event.Alive {
	// 	for portIndex, _ := range event.Ports {
	// 		sd.eventStream <- AddBackendEvent{
	// 			ServiceId: makeServiceId(event.AppId, portIndex),
	// 			BackendId: event.TaskId,
	// 			Hostname:  event.Host,
	// 			Port:      0, //event.Ports[portIndex],
	// 		}
	// 	}
	// } else {
	// 	for portIndex, _ := range event.Ports {
	// 		sd.eventStream <- RemoveBackendEvent{
	// 			ServiceId: makeServiceId(event.AppId, portIndex),
	// 			BackendId: event.TaskId,
	// 		}
	// 	}
	// }
}

func (sd *DiscoveryMarathon) getMarathonApp(appID string) (*marathon.App, error) {
	m, err := marathon.NewService(sd.marathonIP, sd.marathonPort)
	if err != nil {
		return nil, err
	}

	app, err := m.GetApp(appID)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (sd *DiscoveryMarathon) LoadAllApps() {
	var apps []*marathon.App
	var err error
	for {
		apps, err = sd.getAllMarathonApps()
		if err != nil {
			sd.log(fmt.Sprintf("Failed to load all apps. %v", err.Error()))
		} else {
			break
		}
	}

	sd.eventStream <- RestoreFromSnapshotEvent{}

	for _, app := range apps {
		for portIndex, portDef := range app.PortDefinitions {
			serviceId := makeServiceId(app.Id, portIndex)
			switch proto := getApplicationProtocol(app, portIndex); proto {
			case "http":
				log.Printf("Found %q with proto %q", serviceId, proto)
				sd.eventStream <- AddHttpServiceEvent{
					ServiceId: serviceId,
					Hosts:     makeStringArray(portDef.Labels[LB_VHOST_HTTP]),
				}
			case "tcp":
				sd.log("TODO: protocol: \"tcp\"")
			case "udp":
				sd.log("TODO: protocol: \"udp\"")
			default:
				sd.log(fmt.Sprintf("Unhandled protocol: %q", proto))
			}
			for _, task := range app.Tasks {
				if task.IsAlive() {
					sd.eventStream <- AddBackendEvent{
						ServiceId: serviceId,
						BackendId: task.Id,
						Hostname:  task.Host,
						Port:      task.Ports[portIndex]}
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Marathon helpers

func makeServiceId(appId string, portIndex int) string {
	return fmt.Sprintf("%v-%v", appId, portIndex)
}

func getApplicationProtocol(app *marathon.App, portIndex int) string {
	if proto := getHealthCheckProtocol(app, portIndex); len(proto) != 0 {
		return proto
	}

	if proto := getTransportProtocol(app, portIndex); len(proto) != 0 {
		return proto
	}

	return ""
}

func getHealthCheckProtocol(app *marathon.App, portIndex int) string {
	for _, hs := range app.HealthChecks {
		if hs.PortIndex == portIndex {
			if strings.HasPrefix(hs.Protocol, "MESOS_") {
				return strings.ToLower(hs.Protocol[6:])
			} else {
				return strings.ToLower(hs.Protocol)
			}
		}
	}

	return ""
}

func getTransportProtocol(app *marathon.App, portIndex int) string {
	if portIndex < len(app.PortDefinitions) {
		return app.PortDefinitions[portIndex].Protocol // already lower-case
	}

	if app.Container.Docker != nil && portIndex < len(app.Container.Docker.PortMappings) {
		return strings.ToLower(app.Container.Docker.PortMappings[portIndex].Protocol)
	}

	if len(app.PortDefinitions) > 0 {
		return "tcp" // default to TCP if at least one port was exposed (host networking)
	}

	return "" // no ports exposed
}

func (sd *DiscoveryMarathon) getAllMarathonApps() ([]*marathon.App, error) {
	m, err := marathon.NewService(sd.marathonIP, sd.marathonPort)
	if err != nil {
		return nil, fmt.Errorf("Could not create new marathon service. %v", err)
	}

	apps, err := m.GetApps()
	if err != nil {
		return nil, fmt.Errorf("Could not get apps. %v", err)
	}

	return apps, err
}
