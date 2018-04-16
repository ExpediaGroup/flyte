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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestCompleteAction_ShouldFinishAction_WhenItExistsAndIsInPendingState(t *testing.T) {

	//Given
	defer resetActionRepo()
	completedAction := Action{State: State{Value: stateSuccess}, Result: Event{Name: "resultEvent"}}
	actualUpdateAction := Action{}
	calledGet := false
	actionRepo = mockActionRepo{
		get: func(actionId string) (*Action, error) {
			calledGet = true
			if actionId == "existingPendingAction" {
				return &Action{State: State{Value: statePending}}, nil
			}
			return nil, ActionNotFoundErr
		},
		update: func(action Action) error {
			actualUpdateAction = action
			return nil
		},
	}

	//When
	got, err := Pack{Id: "packA"}.CompleteAction("existingPendingAction", Event{Name: "resultEvent"})
	require.NoError(t, err)
	require.NotNil(t, got)
	require.True(t, calledGet)

	//Then
	assert.WithinDuration(t, time.Now(), got.State.Time, 10*time.Second)
	assert.Equal(t, *got, actualUpdateAction)
	got.State.Time = time.Time{}
	got.prevState = State{}
	assert.Equal(t, completedAction, *got)
}

func TestCompleteAction_ShouldSetActionStateToFatalForFatalResult(t *testing.T) {

	//Given
	defer resetActionRepo()
	actionRepo = mockActionRepo{
		get: func(actionId string) (*Action, error) {
			return &Action{State: State{Value: statePending}}, nil
		},
		update: func(action Action) error {
			return nil
		},
	}

	//When
	got, err := Pack{Id: "packA"}.CompleteAction("existingPendingAction", Event{Name: fatalEventName})
	require.NoError(t, err)
	require.NotNil(t, got)

	//Then
	expectedAction := Action{State: State{Value: stateFatal}, Result: Event{Name: fatalEventName}}
	got.State.Time = time.Time{}
	got.prevState = State{}
	assert.Equal(t, expectedAction, *got)
}

func TestCompleteAction_ShouldReturnErrorWhenActionIsInNewState(t *testing.T) {

	defer resetActionRepo()
	actionRepo = mockActionRepo{
		get: func(actionId string) (*Action, error) {
			return &Action{State: State{Value: stateNew}}, nil
		},
	}

	_, err := Pack{Id: "packA"}.CompleteAction("new", Event{Name: "resultEvent"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action is not in PENDING state")
}

func TestCompleteAction_ShouldReturnErrorWhenActionIsInSuccessState(t *testing.T) {

	defer resetActionRepo()
	actionRepo = mockActionRepo{
		get: func(actionId string) (*Action, error) {
			return &Action{State: State{Value: stateSuccess}}, nil
		},
	}

	_, err := Pack{Id: "packA"}.CompleteAction("success", Event{Name: "resultEvent"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action is not in PENDING state")
}

func TestCompleteAction_ShouldReturnErrorWhenActionIsInFatalState(t *testing.T) {

	defer resetActionRepo()
	actionRepo = mockActionRepo{
		get: func(actionId string) (*Action, error) {
			return &Action{State: State{Value: stateFatal}}, nil
		},
	}

	_, err := Pack{Id: "packA"}.CompleteAction("fatal", Event{Name: "resultEvent"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action is not in PENDING state")
}

func TestCompleteAction_ShouldReturnActionNotFoundErrWhenItDoesNotExist(t *testing.T) {

	defer resetActionRepo()
	actionRepo = mockActionRepo{
		get: func(actionId string) (*Action, error) {
			return nil, ActionNotFoundErr
		},
	}

	_, err := Pack{Id: "packA"}.CompleteAction("nonExisting", Event{Name: "resultEvent"})

	assert.EqualError(t, err, ActionNotFoundErr.Error())
}

func TestCompleteAction_ShouldReturnErrorProducedWhileSearchingForAction(t *testing.T) {

	defer resetActionRepo()
	expectedError := errors.New("something went terribly wrong Sir")
	actionRepo = mockActionRepo{
		get: func(actionId string) (*Action, error) {
			return nil, expectedError
		},
	}

	_, err := Pack{Id: "packA"}.CompleteAction("error", Event{Name: "resultEvent"})

	assert.EqualError(t, err, expectedError.Error())
}

func TestCompleteAction_ShouldReturnErrorProducedWhileUpdatingAction(t *testing.T) {

	//Given
	defer resetActionRepo()
	expectedError := errors.New("something went terribly wrong Sir")
	actionRepo = mockActionRepo{
		get: func(actionId string) (*Action, error) {
			return &Action{State: State{Value: statePending}}, nil
		},
		update: func(action Action) error {
			return expectedError
		},
	}

	//When
	_, err := Pack{Id: "packA"}.CompleteAction("error", Event{Name: "resultEvent"})

	//Then
	assert.EqualError(t, err, expectedError.Error())
}

func TestTakeAction_ShouldReturnActionInPendingState_WhenPackHadNewActionWithTheGivenName(t *testing.T) {

	//Given
	defer resetActionRepo()
	calledFindNew := false
	pendingAction := Action{State: State{Value: statePending}}
	actualUpdateAction := Action{}
	actionRepo = mockActionRepo{
		findNew: func(pack Pack, name string) (*Action, error) {
			calledFindNew = true
			if pack.Id == "packA" && name == "specificName" {
				return &Action{State: State{Value: stateNew}}, nil
			}
			return nil, ActionNotFoundErr
		},
		update: func(action Action) error {
			actualUpdateAction = action
			return nil
		},
	}

	//When
	got, err := Pack{Id: "packA"}.TakeAction("specificName")
	require.NoError(t, err)
	require.True(t, calledFindNew)

	//Then
	assert.WithinDuration(t, time.Now(), got.State.Time, 10*time.Second)
	assert.Equal(t, *got, actualUpdateAction)
	got.State.Time = time.Time{}
	got.prevState = State{}
	assert.Equal(t, pendingAction, *got)
}

func TestTakeAction_ShouldReturnActionInPendingState_WhenPackHasAnyNewActionAndNameWasNotSpecified(t *testing.T) {

	//Given
	defer resetActionRepo()
	actionRepo = mockActionRepo{
		findNew: func(pack Pack, name string) (*Action, error) {
			return &Action{State: State{Value: stateNew}}, nil
		},
		update: func(action Action) error {
			return nil
		},
	}

	//When
	got, err := Pack{Id: "packA"}.TakeAction("")
	require.NoError(t, err)

	//Then
	assert.WithinDuration(t, time.Now(), got.State.Time, 10*time.Second)
	got.State.Time = time.Time{}
	got.prevState = State{}
	expectedAction := Action{State: State{Value: statePending}}
	assert.Equal(t, expectedAction, *got)
}

func TestTakeAction_ShouldReturnNilWhenPackDoesNotHaveNewActions(t *testing.T) {

	defer resetActionRepo()
	actionRepo = mockActionRepo{
		findNew: func(pack Pack, name string) (*Action, error) {
			return nil, nil
		},
	}

	got, err := Pack{Id: "packA"}.TakeAction("noNewActions")
	require.NoError(t, err)

	assert.Nil(t, got)
}

func TestTakeAction_ShouldReturnErrorIfItHappensWhileTryingToFindNewActions(t *testing.T) {

	defer resetActionRepo()
	expectedError := errors.New("not juju error again")
	actionRepo = mockActionRepo{
		findNew: func(pack Pack, name string) (*Action, error) {
			return nil, expectedError
		},
	}

	_, err := Pack{Id: "packA"}.TakeAction("")

	assert.EqualError(t, err, expectedError.Error())
}

func TestTakeAction_ShouldReturnErrorForPendingAction(t *testing.T) {

	defer resetActionRepo()
	actionRepo = mockActionRepo{
		findNew: func(pack Pack, name string) (*Action, error) {
			return &Action{State: State{Value: statePending}}, nil
		},
	}

	_, err := Pack{Id: "packA"}.TakeAction("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action is not in NEW state, cannot set to PENDING")
}

func TestTakeAction_ShouldReturnErrorForSuccessAction(t *testing.T) {

	defer resetActionRepo()
	actionRepo = mockActionRepo{
		findNew: func(pack Pack, name string) (*Action, error) {
			return &Action{State: State{Value: stateSuccess}}, nil
		},
	}

	_, err := Pack{Id: "packA"}.TakeAction("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action is not in NEW state, cannot set to PENDING")
}

func TestTakeAction_ShouldReturnErrorForFatalAction(t *testing.T) {

	defer resetActionRepo()
	actionRepo = mockActionRepo{
		findNew: func(pack Pack, name string) (*Action, error) {
			return &Action{State: State{Value: stateFatal}}, nil
		},
	}

	_, err := Pack{Id: "packA"}.TakeAction("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "action is not in NEW state, cannot set to PENDING")
}

// --- mocks & helpers ---

type mockActionRepo struct {
	add            func(a Action) error
	get            func(actionId string) (*Action, error)
	update         func(a Action) error
	findNew        func(p Pack, name string) (*Action, error)
	findCorrelated func(correlationId string) ([]Action, error)
}

func (r mockActionRepo) Add(a Action) error {
	return r.add(a)
}

func (r mockActionRepo) Get(actionId string) (*Action, error) {
	return r.get(actionId)
}

func (r mockActionRepo) Update(a Action) error {
	return r.update(a)
}

func (r mockActionRepo) FindNew(p Pack, name string) (*Action, error) {
	return r.findNew(p, name)
}

func (r mockActionRepo) FindCorrelated(correlationId string) ([]Action, error) {
	return r.findCorrelated(correlationId)
}

func resetActionRepo() { actionRepo = actionMgoRepo{} }
