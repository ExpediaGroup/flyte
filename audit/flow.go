/*
Copyright (C) 2018 Expedia Group.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package audit

import (
	"github.com/ExpediaGroup/flyte/json"
	"time"
)

type Flow struct {
	Name          string            `json:"name" bson:"name"`
	UUID          string            `json:"uuid" bson:"uuid"`
	CorrelationId string            `json:"correlationId" bson:"-"`
	Steps         []Step            `json:"steps" bson:"steps,omitempty"`
	Actions       map[string]Action `json:"actions" bson:"-"`
}

type Step struct {
	Id        string            `json:"id" bson:"id"`
	DependsOn []string          `json:"dependsOn,omitempty" bson:"dependsOn,omitempty"`
	Event     EventDef          `json:"event" bson:"event"`
	Context   map[string]string `json:"context,omitempty" bson:"context,omitempty"`
	Criteria  string            `json:"criteria,omitempty" bson:"criteria,omitempty"`
	Command   Command           `json:"command" bson:"command"`
}

type EventDef struct {
	Name       string            `json:"name" bson:"name"`
	PackName   string            `json:"packName" bson:"packName"`
	PackLabels map[string]string `json:"packLabels,omitempty" bson:"packLabels,omitempty"`
}

type Command struct {
	Name       string            `json:"name" bson:"name"`
	PackName   string            `json:"packName" bson:"packName"`
	PackLabels map[string]string `json:"packLabels,omitempty" bson:"packLabels,omitempty"`
	Input      json.Json         `json:"input" bson:"input,omitempty"`
}

type Action struct {
	Id         string            `json:"id" bson:"_id"`
	Name       string            `json:"name" bson:"name"`
	PackName   string            `json:"packName" bson:"packName"`
	PackLabels map[string]string `json:"packLabels,omitempty" bson:"packLabels,omitempty"`
	Input      json.Json         `json:"input,omitempty" bson:"input,omitempty"`
	State      State             `json:"state" bson:"state"`
	States     []State           `json:"states,omitempty" bson:"states"`

	CorrelationId string `json:"correlationId" bson:"correlationId"`
	FlowName      string `json:"flowName" bson:"flowName"`
	FlowUUID      string `json:"flowUUID" bson:"flowUUID"`
	StepId        string `json:"stepId" bson:"stepId"`

	Context map[string]string `json:"context,omitempty" bson:"context,omitempty"`
	Trigger Event             `json:"trigger" bson:"trigger"`
	Result  Event             `json:"result,omitempty" bson:"result,omitempty"`
}

type Pack struct {
	Id     string            `json:"id" bson:"_id"`
	Name   string            `json:"name" bson:"name"`
	Labels map[string]string `json:"labels,omitempty" bson:"labels,omitempty"`
}

type State struct {
	Value string    `json:"value" bson:"value"`
	Time  time.Time `json:"time" bson:"time"`
}

type Event struct {
	Name       string    `json:"event" bson:"name"`
	Pack       Pack      `json:"pack" bson:"pack"`
	Payload    json.Json `json:"payload,omitempty" bson:"payload,omitempty"`
	CreatedAt  time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
	ReceivedAt time.Time `json:"receivedAt,omitempty" bson:"receivedAt,omitempty"`
}

type Repository interface {
	Get(correlationId string) (*Flow, error)
	Find(filter flowsFilter) ([]Flow, error)
}
