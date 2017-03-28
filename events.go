// This file is part of the "sad" project
//   <http://github.com/christianparpart/sad>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package main

type AddHttpServiceEvent struct {
	ServiceId string
	Host      string
}

type AddBackendEvent struct {
	ServiceId string
	BackendId string
	Hostname  string
	Port      uint
}

type RemoveBackendEvent struct {
	ServiceId string
	BackendId string
}

type LogEvent struct {
	Message string
}
