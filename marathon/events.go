// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package marathon

import "time"

type TaskStatus string

const (
	// TaskStaging means, this task is in staging (such as 'docker pull'-state)
	TaskStaging  = TaskStatus("TASK_STAGING")
	TaskStarting = TaskStatus("TASK_STARTING")
	TaskRunning  = TaskStatus("TASK_RUNNING")
	TaskFinished = TaskStatus("TASK_FINISHED")
	TaskFailed   = TaskStatus("TASK_FAILED")
	TaskKilling  = TaskStatus("TASK_KILLING")
	TaskKilled   = TaskStatus("TASK_KILLED")
	TaskLost     = TaskStatus("TASK_LOST")
)

type IpAddr struct {
	IpAddress string
	Protocol  string
}

/*
{
	"slaveId":"9e1a18f2-011c-44fe-9715-be1cac1d5f41-S8",
	"taskId":"production_lovemag_app.03eb79d1-058d-11e6-a243-72491c981fcc",
	"taskStatus":"TASK_RUNNING",
	"message":"",
	"appId":"/production/lovemag/app",
	"host":"rack2-compute5.datacenter.tld",
	"ipAddresses": [
		{
			"ipAddress":"172.17.0.7",
			"protocol":"IPv4"
		}
	],
	"ports": [47755],
	"version":"2016-04-14T12:52:12.465Z",
	"eventType":"status_update_event",
	"timestamp":"2016-04-18T17:43:10.580Z"
}
*/
type StatusUpdateEvent struct {
	EventType   string     `json:"eventType"`
	Timestamp   time.Time  `json:"timestamp"`
	AppId       string     `json:"appId"`
	TaskId      string     `json:"taskId"`
	SlaveId     string     `json:"slaveId"`
	TaskStatus  TaskStatus `json:"taskStatus"`
	Message     string     `json:"message"`
	Host        string     `json:"host"`
	Ports       []uint     `json:"ports"`
	IpAddresses []IpAddr   `json:"ipAddresses"`
	Version     string     `json:"version"`
}

type HealthStatusChangedEvent struct {
	AppId  string `json:"appId"`
	TaskId string `json:"taskId"`
	Alive  bool   `json:"alive"`
}

type DeploymentInfoEvent struct {
	Plan        DeploymentPlan `json:"plan"`
	CurrentStep DeploymentStep `json:"currentStep"`
	Timestamp   time.Time      `json:"timestamp"`
}

type DeploymentSuccessEvent struct {
	Plan      DeploymentPlan `json:"plan"`
	Timestamp time.Time      `json:"timestamp"`
}

type DeploymentFailedEvent struct {
	Id        string         `json:"id"`
	Plan      DeploymentPlan `json:"plan"`
	Timestamp time.Time      `json:"timestamp"`
}

type DeploymentPlan struct {
	Id       string           `json:"id"`
	Original DeploymentTarget `json:"original"`
	Target   DeploymentTarget `json:"target"`
	Steps    []DeploymentStep `json:"steps"`
	Version  time.Time        `json:"version"`
}

type DeploymentTarget struct {
	Id           string        `json:"id"`
	Apps         []App         `json:"apps"`
	Dependencies []interface{} `json:"dependencies"` // TODO
	Groups       []interface{} `json:"groups"`       // TODO
	Version      time.Time     `json:"version"`
}

type DeploymentStep struct {
	Actions []DeploymentAction `json:"actions"`
}

type DeploymentAction struct {
	Action string `json:"action"`
	App    string `json:"app"`
}

type AppTerminatedEvent struct {
	AppId     string    `json:"appId"`
	Timestamp time.Time `json:"timestamp"`
}

type AddHealthCheckEvent struct {
	AppId       string      `json:"appId"`
	Version     time.Time   `json:"version"`
	HealthCheck HealthCheck `json:"healthCheck"`
	Timestamp   time.Time   `json:"timestamp"`
}

type RemoveHealthCheckEvent struct {
	AppId     string    `json:"appId"`
	Timestamp time.Time `json:"timestamp"`
}
