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

package execution

import (
	"errors"
	"fmt"
	"github.com/ExpediaGroup/flyte/json"
	"time"
)

type Action struct {
	Id         string            `bson:"_id"`
	Name       string            `bson:"name"`
	PackName   string            `bson:"packName"`
	PackLabels map[string]string `bson:"packLabels,omitempty"`
	Input      json.Json         `bson:"input,omitempty"`
	State      State             `bson:"state"`
	prevState  State             `bson:"_"`

	CorrelationId string `bson:"correlationId"`
	FlowName      string `bson:"flowName"`
	FlowUUID      string `bson:"flowUUID"`
	StepId        string `bson:"stepId"`

	Context map[string]string `bson:"context,omitempty"`
	Trigger Event             `bson:"trigger"`
	Result  Event             `bson:"result,omitempty"`
}

func (a *Action) take() error {

	if a.State.Value != stateNew {
		return fmt.Errorf("action is not in %s state, cannot set to %s", stateNew, statePending)
	}
	a.setState(statePending)
	return actionRepo.Update(*a)
}

func (a *Action) finish(e Event) error {

	if a.State.Value != statePending {
		return fmt.Errorf("action is not in %s state", statePending)
	}

	if e.isFatal() {
		a.setState(stateFatal)
	} else {
		a.setState(stateSuccess)
	}
	a.Result = e
	return actionRepo.Update(*a)
}

func (a Action) hasFinished() bool {
	return a.State.Value == stateSuccess || a.State.Value == stateFatal
}

func (a *Action) setState(state string) {
	a.prevState = a.State
	a.State = State{Value: state, Time: time.Now()}
}

type State struct {
	Value string    `bson:"value"`
	Time  time.Time `bson:"time"`
}

const (
	stateNew     = "NEW"
	statePending = "PENDING"
	stateSuccess = "SUCCESS"
	stateFatal   = "FATAL"
)

type Event struct {
	Name    string    `json:"event" bson:"name"`
	Pack    Pack      `json:"pack" bson:"pack"`
	Payload json.Json `json:"payload" bson:"payload,omitempty"`
}

func (e Event) isFatal() bool {
	return e.Name == fatalEventName
}

const fatalEventName = "FATAL"

type ActionRepository interface {
	Add(action Action) error
	Get(actionId string) (*Action, error)
	Update(action Action) error
	FindNew(pack Pack, name string) (*Action, error)
	FindCorrelated(correlationId string) ([]Action, error)
}

var actionRepo ActionRepository = actionMgoRepo{}

var ActionNotFoundErr = errors.New("action not found")
