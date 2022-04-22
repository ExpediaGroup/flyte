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
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestFlowHandleEvent_ShouldExecuteAllCandidateSteps(t *testing.T) {

	defer resetStepExecutor()
	rec := setupStepExecutorWithAction(nil)

	defer resetActionRepo()
	addCounter := 0
	actionRepo = mockActionRepo{
		add: func(a Action) error {
			addCounter++
			return nil

		},
	}

	defer resetAuditRepo()
	auditCounter := 0
	auditRepo = mockAuditRepo{
		add: func(a Action) error {
			auditCounter++
			return nil

		},
	}

	candidateA := newStepT("candidateA", "eventOK", "packOK")
	nonCandidate := newStepT("nonCandidate", "eventNotOK", "packOK")
	candidateB := newStepT("candidateB", "eventOK", "packOK")
	flow := newFlowT(candidateA, nonCandidate, candidateB)

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})

	assert.Len(t, rec.calls, 2)
	assert.Contains(t, rec.steps(), candidateA)
	assert.Contains(t, rec.steps(), candidateB)

	assert.Len(t, flow.actions, 2)
	assert.Equal(t, Action{Id: candidateA.Id, StepId: candidateA.Id}, flow.actions[candidateA.Id])
	assert.Equal(t, Action{Id: candidateB.Id, StepId: candidateB.Id}, flow.actions[candidateB.Id])

	assert.Equal(t, 2, addCounter)
	assert.Equal(t, 2, auditCounter)
}

func TestFlowHandleEvent_ShouldCreateActionWhichIncludesFlowName(t *testing.T) {
	setupStepExecutorWithAction(nil)
	defer resetStepExecutor()

	actionRepo = mockActionRepo{
		add: func(a Action) error {
			return nil
		},
	}
	defer resetActionRepo()

	defer resetAuditRepo()
	auditRepo = mockAuditRepo{
		add: func(a Action) error {
			return nil
		},
	}

	stepA := newStepT("stepA", "eventOK", "packOK")
	flow := Flow{
		Name:    "flowA",
		Steps:   []Step{stepA},
		actions: map[string]Action{},
	}

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})

	assert.Len(t, flow.actions, 1)
	assert.Equal(t, Action{Id: stepA.Id, StepId: stepA.Id, FlowName: flow.Name}, flow.actions[stepA.Id])
}

func TestFlowHandleEvent_ShouldLogCreatedActions(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelInfo)

	defer resetStepExecutor()
	setupStepExecutorWithAction(nil)

	defer resetActionRepo()
	actionRepo = mockActionRepo{
		add: func(a Action) error {
			return nil
		},
	}
	defer resetAuditRepo()
	auditRepo = mockAuditRepo{
		add: func(a Action) error {
			return nil
		},
	}
	flow := newFlowT(newStepT("stepA", "eventOK", "packOK"))

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Action has been created actionId=stepA", logMessages[0].Message)
}

func TestFlowHandleEvent_ShouldSkipStepWhenEventNameDoesNotMatch(t *testing.T) {

	defer resetStepExecutor()
	rec := setupStepExecutor(nil, nil)

	flow := newFlowT(newStepT("stepA", "eventA", "packOK"))

	flow.HandleEvent(Event{Name: "eventNotOK", Pack: Pack{Name: "packOK"}})

	assert.Len(t, rec.calls, 0)
}

func TestFlowHandleEvent_ShouldSkipStepWhenEventPackNameDoesNotMatch(t *testing.T) {

	defer resetStepExecutor()
	rec := setupStepExecutor(nil, nil)

	flow := newFlowT(newStepT("stepA", "eventOK", "packA"))

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packNotOK"}})

	assert.Len(t, rec.calls, 0)
}

func TestFlowHandleEvent_ShouldSkipStepsWhenThereIsAlreadyActionForThatStep(t *testing.T) {

	defer resetStepExecutor()
	rec := setupStepExecutor(nil, nil)

	flow := newFlowT(newStepT("stepWithExistingFlowAction", "eventOK", "packOK"))
	flow.actions = map[string]Action{"stepWithExistingFlowAction": {}}

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})

	assert.Len(t, rec.calls, 0)
}

func TestFlowHandleEvent_ShouldProceedWithStepExecutionWhenItDoesNotDependOnOtherSteps(t *testing.T) {

	defer resetStepExecutor()
	rec := setupStepExecutor(nil, nil)

	stepWithoutDependsOn := newStepT("withoutDependsOn", "eventOK", "packOK")
	flow := newFlowT(stepWithoutDependsOn)

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})

	assert.Len(t, rec.calls, 1)
	assert.Contains(t, rec.steps(), stepWithoutDependsOn)
}

func TestFlowHandleEvent_ShouldProceedWithStepExecutionWhenDependsOnIsSatisfied(t *testing.T) {

	defer resetStepExecutor()
	rec := setupStepExecutor(nil, nil)

	stepWithDepends := newStepT("withDependsOn", "eventOK", "packOK")
	stepWithDepends.DependsOn = []string{"stepA"}
	flow := newFlowT(stepWithDepends)
	flow.actions["stepA"] = Action{State: State{Value: stateSuccess}}

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})

	assert.Len(t, rec.calls, 1)
	assert.Contains(t, rec.steps(), stepWithDepends)
}

func TestFlowHandleEvent_ShouldSkipStepWhenDependsOnForAStepIsNotSatisfied(t *testing.T) {

	defer resetStepExecutor()
	rec := setupStepExecutor(nil, nil)

	step := newStepT("notSatisfiedDependsOn", "eventOK", "packOK")
	step.DependsOn = []string{"notFinishedStep"}
	flow := newFlowT(step)
	flow.actions["notFinishedStep"] = Action{State: State{Value: stateNew}}

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})

	assert.Len(t, rec.calls, 0)
}

func TestFlowHandleEvent_ShouldProduceNothingWhenStepExecutesToNil(t *testing.T) {

	defer resetStepExecutor()
	rec := setupStepExecutor(nil, nil)

	defer resetActionRepo()
	addNotCalled := true
	actionRepo = mockActionRepo{
		add: func(a Action) error {
			addNotCalled = false
			return nil
		},
	}

	nilActionStep := newStepT("nilAction", "eventOK", "packOK")
	flow := newFlowT(nilActionStep)

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})

	assert.Len(t, rec.calls, 1)
	assert.Contains(t, rec.steps(), nilActionStep)
	assert.True(t, addNotCalled)
	assert.True(t, len(flow.actions) == 0)
}

func TestFlowHandleEvent_ShouldLogErrorAndContinueWhenStepExecutionReturnsError(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetStepExecutor()
	var stepResolverExecs []Step
	stepExecutor = func(s Step, e Event, parentCtx map[string]string) (*Action, error) {
		stepResolverExecs = append(stepResolverExecs, s)
		switch s.Id {
		case "returnError":
			return nil, errors.New("are you for real")
		default:
			return nil, nil
		}
	}

	errorStep := newStepT("returnError", "eventOK", "packOK")
	candidateStep := newStepT("candidateStep", "eventOK", "packOK")
	flow := newFlowT(errorStep, candidateStep)

	flow.HandleEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Error handling flow= step=returnError: are you for real", logMessages[0].Message)

	assert.Len(t, stepResolverExecs, 2)
}

// --- mocks & helpers ---

func newStepT(id, eventName, eventPackName string) Step {
	return Step{
		Id:    id,
		Event: EventDef{Name: eventName, PackName: eventPackName},
	}
}

func newFlowT(steps ...Step) Flow {
	return Flow{
		Steps:   steps,
		actions: map[string]Action{},
	}
}

func setupStepExecutor(a *Action, expectedErr error) *stepExecutorRec {
	rec := &stepExecutorRec{
		calls: []stepExecutorCall{},
	}
	stepExecutor = func(s Step, e Event, ctx map[string]string) (*Action, error) {
		rec.addCall(s, e, ctx)
		return a, expectedErr
	}
	return rec
}

func setupStepExecutorWithAction(expectedErr error) *stepExecutorRec {
	rec := &stepExecutorRec{
		calls: []stepExecutorCall{},
	}
	stepExecutor = func(s Step, e Event, ctx map[string]string) (*Action, error) {
		rec.addCall(s, e, ctx)
		return &Action{Id: s.Id}, expectedErr
	}
	return rec
}

type stepExecutorRec struct {
	calls []stepExecutorCall
}

type stepExecutorCall struct {
	step  Step
	event Event
	ctx   map[string]string
}

func (r *stepExecutorRec) addCall(s Step, e Event, ctx map[string]string) {
	if r.calls == nil {
		r.calls = []stepExecutorCall{}
	}
	r.calls = append(r.calls, stepExecutorCall{s, e, ctx})
}

func (r stepExecutorRec) steps() []Step {
	var steps []Step
	if r.calls != nil {
		for _, c := range r.calls {
			steps = append(steps, c.step)
		}
	}
	return steps
}

func resetStepExecutor() { stepExecutor = executeStep }
