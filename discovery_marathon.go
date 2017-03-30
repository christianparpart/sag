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
	marathonIP    net.IP
	marathonPort  uint
	portsMapCache map[string]int
	sse           *EventSource
	eventStream   chan<- interface{}
}

func NewDiscoveryMarathon(host net.IP, port uint, reconnectDelay time.Duration, eventStream chan<- interface{}) *DiscoveryMarathon {
	url := fmt.Sprintf("http://%v:%v/v2/events", host, port)
	sse := NewEventSource(url, reconnectDelay)

	sd := &DiscoveryMarathon{
		marathonIP:    host,
		marathonPort:  port,
		portsMapCache: make(map[string]int),
		sse:           sse,
		eventStream:   eventStream}

	sse.OnOpen = sd.onOpen
	sse.OnError = sd.onError
	sse.AddEventListener("status_update_event", sd.status_update_event)
	sse.AddEventListener("health_status_changed_event", sd.health_status_changed_event)
	sse.AddEventListener("instance_health_changed_event", sd.instance_health_changed_event)
	sse.AddEventListener("app_terminated_event", sd.app_terminated_event)
	sse.AddEventListener("deployment_info", func(data string) {})
	sse.AddEventListener("deployment_step_success", func(data string) {})
	sse.AddEventListener("deployment_success", func(data string) {})
	sse.AddEventListener("deployment_failed", func(data string) {})
	sse.AddEventListener("event_stream_attached", func(data string) {})
	sse.AddEventListener("event_stream_detached", func(data string) {})
	sse.AddEventListener("api_post_event", func(data string) {})
	sse.AddEventListener("add_health_check_event", func(data string) {})
	sse.AddEventListener("remove_health_check_event", func(data string) {})
	sse.AddEventListener("group_change_success", func(data string) {})
	sse.AddEventListener("group_change_failed", func(data string) {})
	sse.AddEventListener("instance_changed_event", func(data string) {})
	sse.AddEventListener("failed_health_check_event", func(data string) {})
	sse.OnMessage = func(event, data string) {
		log.Printf("SSE: unknown event %v: %v", event, data)
	}

	return sd
}

func (sd *DiscoveryMarathon) String() string {
	return fmt.Sprintf("DiscoveryMarathon<%v>", sd.sse.Url)
}

func (sd *DiscoveryMarathon) Run() {
	log.Printf("Starting Marathon SSE event stream %v:%v", sd.marathonIP, sd.marathonPort)
	sd.sse.RunForever()
}

func (sd *DiscoveryMarathon) Shutdown() {
	sd.sse.Close()
}

func (sd *DiscoveryMarathon) RefreshAllApps() {
	var apps []*marathon.App
	var err error
	for {
		apps, err = sd.getAllMarathonApps()
		if err != nil {
			log.Printf("Failed to load all apps. %v", err.Error())
		} else {
			break
		}
	}

	sd.eventStream <- RestoreFromSnapshotEvent{}

	for _, app := range apps {
		sd.ensureAppIsPropagated(app)
	}

	for _, app := range apps {
		for portIndex, _ := range app.PortDefinitions {
			serviceId := makeServiceId(app.Id, portIndex)
			for _, task := range app.Tasks {
				sd.eventStream <- AddBackendEvent{
					ServiceId: serviceId,
					BackendId: task.Id,
					Hostname:  task.Host,
					Port:      task.Ports[portIndex],
					Alive:     task.IsAlive(),
				}
			}
		}
	}
}

func (sd *DiscoveryMarathon) onOpen() {
	log.Printf("SSE stream connected")
	sd.RefreshAllApps()
}

func (sd *DiscoveryMarathon) onError(message string) {
	log.Printf("SSE failure. %v", message)
}

func (sd *DiscoveryMarathon) status_update_event(data string) {
	var event marathon.StatusUpdateEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		log.Printf("Failed to unmarshal status_update_event. %v", err)
		log.Printf("status_update_event: %+v\n", data)
	}

	switch event.TaskStatus {
	case marathon.TaskRunning:
		// add task iff no health checks are available
		app, _ := sd.getMarathonApp(event.AppId)
		if app != nil && len(app.HealthChecks) == 0 {
		}
		sd.ensureAppIsPropagated(app)
		for portIndex, _ := range app.PortDefinitions {
			serviceId := makeServiceId(event.AppId, portIndex)
			sd.eventStream <- AddBackendEvent{
				ServiceId: serviceId,
				BackendId: event.TaskId,
				Hostname:  event.Host,
				Port:      event.Ports[portIndex],
				Alive:     len(app.HealthChecks) == 0,
			}
			// XXX we consider the backend already alive when there are no
			// health-checks defined but ports defined.
			// If there are health checks defined, Alive is initially set to false, and
			// a health_status_changed_event to enable itwill follow up to enable it.
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
		log.Printf("Failed to unmarshal health_status_changed_event. %v", err)
		return
	}

	if maxPorts, ok := sd.portsMapCache[event.AppId]; ok {
		for portIndex := 0; portIndex < maxPorts; portIndex++ {
			sd.eventStream <- HealthStatusChangedEvent{
				ServiceId: makeServiceId(event.AppId, portIndex),
				BackendId: event.TaskId,
				Alive:     bool(event.Alive),
			}
		}
	}
}

func (sd *DiscoveryMarathon) instance_health_changed_event(data string) {
	var event marathon.InstanceHealthChangedEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		log.Printf("Failed to unmarshal instance_health_changed_event. %v\n", err)
		return
	}

	// log.Printf("%v: %+v", event.EventType, event)
}

func (sd *DiscoveryMarathon) app_terminated_event(data string) {
	var event marathon.AppTerminatedEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		log.Printf("Failed to unmarshal app_terminated_event. %v\n", err)
		return
	}

	log.Printf("Application terminated. %v", event.AppId)
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

func (sd *DiscoveryMarathon) ensureAppIsPropagated(app *marathon.App) {
	sd.portsMapCache[app.Id] = len(app.PortDefinitions)

	for portIndex, portDef := range app.PortDefinitions {
		serviceId := makeServiceId(app.Id, portIndex)
		proto := getApplicationProtocol(app, portIndex)
		switch proto {
		case "http":
			sd.eventStream <- AddHttpServiceEvent{
				ServiceId:   serviceId,
				ServicePort: portDef.Port,
				Hosts:       makeStringArray(portDef.Labels[LB_VHOST_HTTP]),
			}
		case "tcp":
			sd.eventStream <- AddTcpServiceEvent{
				ServiceId:     serviceId,
				ServicePort:   portDef.Port,
				ProxyProtocol: Atoi(portDef.Labels[LB_PROXY_PROTOCOL], 0),
				AcceptProxy:   MakeBool(portDef.Labels[LB_ACCEPT_PROXY]),
			}
		case "udp":
			sd.eventStream <- AddUdpServiceEvent{
				ServiceId:   serviceId,
				ServicePort: portDef.Port,
			}
		default:
			log.Printf("Unhandled protocol: %q", proto)
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
