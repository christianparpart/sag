// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

type SchedulingAlgorithm string

const (
	SchedulerRoundRobin = SchedulingAlgorithm("round-robin")
	SchedulerLeastLoad  = SchedulingAlgorithm("least-load")
	SchedulerChance     = SchedulingAlgorithm("chance")
)

type RestoreFromSnapshotEvent struct {
}

type AddUdpServiceEvent struct {
	ServiceId   string
	ServicePort uint
	Scheduler   SchedulingAlgorithm
}

type AddTcpServiceEvent struct {
	ServiceId     string
	ServicePort   uint
	Scheduler     SchedulingAlgorithm
	ProxyProtocol int  // ProxyProtocol version to pass to upstream (0=disabled, 1=v1, 2=v2)
	AcceptProxy   bool // AcceptProxy indicates whether or not to parse proxy header from clients
}

type AddHttpServiceEvent struct {
	ServiceId   string
	ServicePort uint
	Scheduler   SchedulingAlgorithm
	Hosts       []string
}

type AddBackendEvent struct {
	ServiceId string
	BackendId string
	Hostname  string
	Port      uint
	Capacity  int
	Alive     bool
}

type RemoveBackendEvent struct {
	ServiceId string
	BackendId string
}

type HealthStatusChangedEvent struct {
	ServiceId string
	BackendId string
	Alive     bool
}

type LogEvent struct {
	Message string
}
