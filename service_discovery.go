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

type ServiceDiscovery interface {
	Run()
	Shutdown()
}

type MarathonSD struct {
	marathonIP   net.IP
	marathonPort int
	sse          *EventSource
	eventStream  chan<- interface{}
}

func NewMarathonSD(host net.IP, port int, reconnectDelay time.Duration, eventStream chan<- interface{}) *MarathonSD {
	url := fmt.Sprintf("http://%v:%v/v2/events", host, port)
	sse := NewEventSource(url, reconnectDelay)

	sd := &MarathonSD{
		marathonIP:   host,
		marathonPort: port,
		sse:          sse,
		eventStream:  eventStream}

	sse.OnOpen = sd.onOpen
	sse.OnError = sd.onError
	sse.AddEventListener("status_update_event", sd.status_update_event)
	sse.AddEventListener("health_status_changed_event", sd.health_status_changed_event)

	return sd
}

func (sd *MarathonSD) String() string {
	return fmt.Sprintf("MarathonSD<%v>", sd.sse.Url)
}

func (sd *MarathonSD) Run() {
	sd.log("Starting")
	sd.sse.RunForever()
}

func (sd *MarathonSD) Shutdown() {
	sd.sse.Close()
}

func (sd *MarathonSD) log(msg string) {
	sd.eventStream <- LogEvent{
		Message: fmt.Sprintf("marathon(%v): %v", sd.sse.Url, msg)}
}

func (sd *MarathonSD) onOpen() {
	sd.log("Connected")
}

func (sd *MarathonSD) onError(message string) {
	sd.log("SSE failure. " + message)
}

func (sd *MarathonSD) status_update_event(data string) {
	var event marathon.StatusUpdateEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		sd.log(fmt.Sprintf("Failed to unmarshal status_update_event. %v", err))
		sd.log(fmt.Sprintf("status_update_event: %+v\n", data))
	} else {
		sd.statusUpdateEvent(&event)
	}
}

func (sd *MarathonSD) statusUpdateEvent(event *marathon.StatusUpdateEvent) {
	switch event.TaskStatus {
	case marathon.TaskRunning:
		// TODO add task, iff no health checks are available
		for i, _ := range event.Ports {
			sd.eventStream <- AddBackendEvent{
				ServiceId: event.AppId,
				BackendId: event.TaskId,
				Hostname:  event.Host,
				Port:      event.Ports[i]}
		}
	case marathon.TaskFinished, marathon.TaskFailed, marathon.TaskKilling, marathon.TaskKilled, marathon.TaskLost:
		// TODO remove task
		sd.log(fmt.Sprintf("statusUpdateEvent: remove task %+v", *event))
	}
}

func (sd *MarathonSD) health_status_changed_event(data string) {
	var event marathon.HealthStatusChangedEvent
	err := json.Unmarshal([]byte(data), &event)
	if err != nil {
		sd.log(fmt.Sprintf("Failed to unmarshal health_status_changed_event. %v\n", err))
	} else {
		sd.healthStatusChangedEvent(&event)
	}
}

func (sd *MarathonSD) healthStatusChangedEvent(event *marathon.HealthStatusChangedEvent) {
	// TODO: add/remove tasks
	if event.Alive {
		sd.log(fmt.Sprintf("healthStatusChangedEvent: %+v", *event))
	} else {
		sd.log(fmt.Sprintf("healthStatusChangedEvent: %+v", *event))
	}
}
