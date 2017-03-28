package main

type AddBackendEvent struct {
	ServiceId string
	BackendId string
	Hostname  string
	Port      int
}

type RemoveBackendEvent struct {
	ServiceId string
	BackendId string
}

type LogEvent struct {
	Message string
}
