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

package flow

import (
	"errors"
	"github.com/HotelsDotCom/flyte/json"
)

// Flow uses two different collections, one contains latest flows (one per name)
// and the the other is history collection containing all flows that have been added.
// Because mongo does NOT allow to update `_id` field we have to use custom `uuid` field, so we can update this field
// in flow collection to make replacing latest flow atomic. This uuid is the same in both collections for the latest flow.
type Flow struct {
	UUID        string `json:"-" bson:"uuid"`
	Name        string `json:"name" bson:"name"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Steps       []Step `json:"steps,omitempty" bson:"steps,omitempty"`
}

type Step struct {
	Id        string            `json:"id,omitempty" bson:"id,omitempty"`
	DependsOn []string          `json:"dependsOn,omitempty" bson:"dependsOn,omitempty"`
	Event     Event             `json:"event" bson:"event"`
	Context   map[string]string `json:"context,omitempty" bson:"context,omitempty"`
	Criteria  string            `json:"criteria,omitempty" bson:"criteria,omitempty"`
	Command   Command           `json:"command" bson:"command"`
}

type Event struct {
	Name       string            `json:"name" bson:"name"`
	PackName   string            `json:"packName" bson:"packName"`
	PackLabels map[string]string `json:"packLabels,omitempty" bson:"packLabels,omitempty"`
}

type Command struct {
	Name       string            `json:"name" bson:"name"`
	PackName   string            `json:"packName" bson:"packName"`
	PackLabels map[string]string `json:"packLabels,omitempty" bson:"packLabels,omitempty"`
	Input      json.Json         `json:"input" bson:"input"`
}

type Repository interface {
	Add(flow Flow) error
	Remove(name string) error
	Get(name string) (*Flow, error)
	FindAll() ([]Flow, error)
}

var FlowNotFoundErr = errors.New("flow not found")
