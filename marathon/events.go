// This file is part of the "sag" project
//   <http://github.com/christianparpart/sag>
//   (c) 2017 Christian Parpart <christian@parpart.family>
//
// Licensed under the MIT License (the "License"); you may not use this
// file except in compliance with the License. You may obtain a copy of
// the License at: http://opensource.org/licenses/MIT

package marathon

import "time"
import "log"

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

type HealthStatus bool

const Healthy HealthStatus = true
const Unhealthy HealthStatus = false

func (hs HealthStatus) String() string {
	switch hs {
	case Healthy:
		return "Healthy"
	case Unhealthy:
		return "Unhealthy"
	default:
		log.Panic("The weird happend")
		return ""
	}
}

type GenericEvent struct {
	EventType string    `json:"eventType"` // "instance_changed_event"
	Timestamp time.Time `json:"timestamp"` // "2017-03-30T10:28:45.822Z"
}

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
	AppId  string       `json:"appId"`
	TaskId string       `json:"taskId"`
	Alive  HealthStatus `json:"alive"`
}

type DeploymentInfoEvent struct {
	*GenericEvent
	Plan        DeploymentPlan `json:"plan"`
	CurrentStep DeploymentStep `json:"currentStep"`
}

type DeploymentSuccessEvent struct {
	*GenericEvent
	Plan DeploymentPlan `json:"plan"`
}

type DeploymentFailedEvent struct {
	*GenericEvent
	Id   string         `json:"id"`
	Plan DeploymentPlan `json:"plan"`
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
	*GenericEvent
	AppId string `json:"appId"`
}

type AddHealthCheckEvent struct {
	*GenericEvent
	AppId       string      `json:"appId"`
	Version     time.Time   `json:"version"`
	HealthCheck HealthCheck `json:"healthCheck"`
}

type RemoveHealthCheckEvent struct {
	*GenericEvent
	AppId string `json:"appId"`
}

type InstanceChangedEvent struct {
	*GenericEvent
	InstanceId     string    `json:"instanceId"`     // "christian-test1.marathon-a64785da-1533-11e7-850f-02421a280e7f"
	Condition      string    `json:"condition"`      // "Failed", "Created", "Running", "Killed"
	RunSpecId      string    `json:"runSpecId"`      // "/christian-test1"
	AgentId        string    `json:"agentId"`        // "15c09a47-86e9-4177-856c-ede35ec02af8-S62"
	Host           string    `json:"host"`           // "somehost123"
	RunSpecVersion time.Time `json:"runSpecVersion"` // "2017-03-30T10:27:36.901Z"
}

type InstanceHealthChangedEvent struct {
	*GenericEvent
	InstanceId     string       `json:"instanceId"`
	RunSpecId      string       `json:"runSpecId"`
	Health         HealthStatus `json:"health"`
	RunSpecVersion time.Time    `json:"runSpecVersion"`
}

//  failed_health_check_event:
type FailedHealthCheckEvent struct {
	*GenericEvent
	AppId       string      `json:"appId"`      // "/christian-test1"
	InstanceId  string      `json:"instanceId"` // "christian-test1.marathon-aa2ae852-1538-11e7-850f-02421a280e7f",
	HealthCheck HealthCheck `json:"healthCheck"`
}

type GroupChangeFailed struct {
	*GenericEvent
	GroupId string    `json:"groupId"` // "/infra/monitoring/alertmanager"
	Reason  string    `json:"reason"`  // "App is locked by one or more deployments. Override with the option '?force=true'. View details at '/v2/deployments/<DEPLOYMENT_ID>'."
	Version time.Time `json:"version"`
}

type UnknownInstanceTerminatedEvent struct {
	*GenericEvent
	InstanceId string `json:"instanceId"` // "production_infra_docker-registry.marathon-0b35086b-10b1-11e7-a6dd-02429399af65"
	RunSpecId  string `json:"runSpecId"`  // "/production/infra/docker-registry"
	Condition  string `json:"condition"`  // "Killed"
}
