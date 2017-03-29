// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

type RestoreFromSnapshotEvent struct {
}

type AddHttpServiceEvent struct {
	ServiceId string
	Hosts     []string
}

type AddBackendEvent struct {
	ServiceId string
	BackendId string
	Hostname  string
	Port      uint
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
