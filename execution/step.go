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
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"github.com/HotelsDotCom/flyte/collections"
	"github.com/HotelsDotCom/flyte/json"
	"github.com/HotelsDotCom/flyte/template"
	"strconv"
	"time"
)

type Step struct {
	Id        string            `bson:"id,omitempty"`
	DependsOn []string          `bson:"dependsOn,omitempty"`
	Event     EventDef          `bson:"event"`
	Context   map[string]string `bson:"context,omitempty"`
	Criteria  string            `bson:"criteria,omitempty"`
	Command   Command           `bson:"command"`
}

type EventDef struct {
	Name       string            `bson:"name"`
	PackName   string            `bson:"packName"`
	PackLabels map[string]string `bson:"packLabels,omitempty"`
}

type Command struct {
	Name       string            `bson:"name"`
	PackName   string            `bson:"packName"`
	PackLabels map[string]string `bson:"packLabels,omitempty"`
	Input      json.Json         `bson:"input,omitempty"`
}

func (s Step) Execute(e Event, parentCtx map[string]string) (*Action, error) {
	return stepExecutor(s, e, parentCtx)
}

var stepExecutor = executeStep

func executeStep(s Step, e Event, parentCtx map[string]string) (*Action, error) {
	ctx, err := s.resolveContext(e, parentCtx)

	if err != nil {
		return nil, err
	}

	if match, err := s.matchesEvent(e, ctx); err != nil || !match {
		return nil, err
	}

	if criteriaMet, err := s.isCriteriaMet(e, ctx); err != nil || !criteriaMet {
		return nil, err
	}

	a, err := s.Command.createAction(e, ctx)
	if a != nil {
		a.StepId = s.Id
	}
	return a, err
}

func (s Step) resolveContext(e Event, parentCtx map[string]string) (map[string]string, error) {

	resolvedCtx, err := template.Resolve(s.Context, templateContext(e, parentCtx))
	if err != nil {
		return nil, fmt.Errorf("error resolving context with event=%+v and ctx=%v: %v", e, parentCtx, err)
	}
	return collections.Merge(parentCtx, resolvedCtx.(map[string]string)), nil
}

func (s Step) matchesEvent(e Event, ctx map[string]string) (bool, error) {

	if s.Event.Name != e.Name || s.Event.PackName != e.Pack.Name {
		return false, nil
	}

	packLabels, err := resolveLabels(s.Event.PackLabels, e, ctx)
	if err != nil {
		return false, err
	}

	return collections.ContainsAll(e.Pack.Labels, packLabels), nil
}

func (s Step) isCriteriaMet(e Event, ctx map[string]string) (bool, error) {

	if s.Criteria == "" {
		return true, nil
	}
	criteria, err := template.Resolve(s.Criteria, templateContext(e, ctx))
	if err != nil {
		return false, fmt.Errorf("error resolving criteria with event=%+v and ctx=%v: %v", e, ctx, err)
	}
	return strconv.ParseBool(criteria.(string))
}

func (c Command) createAction(e Event, ctx map[string]string) (*Action, error) {

	packLabels, err := resolveLabels(c.PackLabels, e, ctx)
	if err != nil {
		return nil, err
	}

	input, err := c.resolveInput(e, ctx)
	if err != nil {
		return nil, err
	}

	return &Action{
		Id:         bson.NewObjectId().Hex(),
		Name:       c.Name,
		PackName:   c.PackName,
		PackLabels: packLabels,
		Input:      input,
		State:      State{Value: stateNew, Time: time.Now()},
		Trigger:    e,
		Context:    ctx,
	}, nil
}

func (c Command) resolveInput(e Event, ctx map[string]string) (json.Json, error) {

	input, err := template.Resolve(c.Input, templateContext(e, ctx))
	if err != nil {
		return nil, fmt.Errorf("error resolving command input with event=%+v and ctx=%v: %v", e, ctx, err)
	}
	return input, nil
}

func resolveLabels(labelsTmpl map[string]string, e Event, ctx map[string]string) (map[string]string, error) {
	labels, err := template.Resolve(labelsTmpl, templateContext(e, ctx))
	if err != nil {
		return nil, fmt.Errorf("error resolving pack labels with event=%+v and ctx=%v: %v", e, ctx, err)
	}
	return labels.(map[string]string), nil
}

func templateContext(event interface{}, ctx map[string]string) template.Context {
	context := make(map[string]interface{})
	for k, v := range ctx {
		context[k] = v
	}

	return template.Context{
		"Event":   event,
		"Context": context,
	}
}
