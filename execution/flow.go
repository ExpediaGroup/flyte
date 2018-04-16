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
	"github.com/HotelsDotCom/go-logger"
)

type Flow struct {
	UUID  string `bson:"uuid"`
	Name  string `bson:"name"`
	Steps []Step `bson:"steps,omitempty"`

	correlationId string            `bson:"-"`
	context       map[string]string `bson:"-"`
	actions       map[string]Action `bson:"-"`
}

func (f *Flow) HandleEvent(e Event) {
	flowEventHandler(f, e)
}

var flowEventHandler = flowEventHandlerFn

func flowEventHandlerFn(f *Flow, e Event) {

	for _, step := range f.candidateSteps(e) {

		action, err := step.Execute(e, f.context)
		if err != nil {
			logger.Errorf("Error handling flow=%s step=%s: %v", f.UUID, step.Id, err)
			continue
		}

		if action != nil {
			if err := f.addAction(step.Id, *action); err != nil {
				logger.Errorf("Error saving action=%+v: %v", action, err)
			} else {
				logger.Infof("Action has been created actionId=%s", action.Id)
				logger.Debugf("action=%+v", action)
			}
		}
	}
}

func (f *Flow) addAction(stepId string, a Action) error {
	a.CorrelationId = f.correlationId
	a.FlowUUID = f.UUID
	a.FlowName = f.Name
	a.StepId = stepId

	if err := actionRepo.Add(a); err != nil {
		return err
	}
	f.actions[stepId] = a
	return nil
}

func (f *Flow) candidateSteps(e Event) []Step {
	var steps []Step
	for _, step := range f.Steps {
		if f.isStepCandidateForExecution(step, e) {
			steps = append(steps, step)
		}
	}
	return steps
}

func (f Flow) isStepCandidateForExecution(step Step, e Event) bool {
	return step.Event.Name == e.Name &&
		step.Event.PackName == e.Pack.Name &&
		!f.hasActionForStep(step.Id) &&
		f.isDependsOnSatisfied(step)
}

func (f Flow) hasActionForStep(stepId string) bool {
	if f.actions == nil {
		return false
	}
	_, ok := f.actions[stepId]
	return ok
}

func (f Flow) isDependsOnSatisfied(step Step) bool {
	if len(step.DependsOn) == 0 {
		return true
	}
	for _, stepId := range step.DependsOn {
		// TODO: currently depends_on is an "OR" operation i.e. {"dependsOn" : ["hipchat_start", "jira_start"] }
		// means that EITHER "hipchat_start" OR "jira_start" must have happened, not necessarily both
		if f.hasFinishedAction(stepId) {
			return true
		}
	}
	return false
}

func (f Flow) hasFinishedAction(stepId string) bool {
	if f.actions == nil {
		return false
	}
	if action, ok := f.actions[stepId]; ok {
		if action.hasFinished() {
			return true
		}
	}
	return false
}
