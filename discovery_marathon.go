// This file is part of the "sad" project
//   <http://github.com/christianparpart/sad>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/dawanda/go-mesos/marathon"
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

func makeServiceId(app *marathon.App, portIndex int) string {
	return fmt.Sprintf("%v-%v-%v", app.Id, portIndex,
		app.PortDefinitions[portIndex].Port)
}

func (sd *DiscoveryMarathon) statusUpdateEvent(event *marathon.StatusUpdateEvent) {
	switch event.TaskStatus {
	case marathon.TaskRunning:
		// add task iff no health checks are available
		app, _ := sd.getMarathonApp(event.AppId)
		if app != nil && len(app.HealthChecks) == 0 {
			for portIndex, portDef := range app.PortDefinitions {
				serviceId := makeServiceId(app, portIndex)
				sd.eventStream <- AddHttpServiceEvent{
					ServiceId: serviceId,
					Host:      portDef.Labels[LB_VHOST_HTTP],
				}
				sd.eventStream <- AddBackendEvent{
					ServiceId: serviceId,
					BackendId: event.TaskId,
					Hostname:  event.Host,
					Port:      portDef.Port}
			}
		}
	case marathon.TaskFinished, marathon.TaskFailed, marathon.TaskKilling, marathon.TaskKilled, marathon.TaskLost:
		app, _ := sd.getMarathonApp(event.AppId)
		for portIndex, _ := range app.PortDefinitions {
			sd.eventStream <- RemoveBackendEvent{
				ServiceId: makeServiceId(app, portIndex),
				BackendId: event.TaskId}
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
	// TODO: add/remove tasks
	if event.Alive {
		sd.log(fmt.Sprintf("healthStatusChangedEvent: %+v", *event))
	} else {
		sd.log(fmt.Sprintf("healthStatusChangedEvent: %+v", *event))
	}
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
