// +build integration

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
	"github.com/ExpediaGroup/flyte/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

func TestAdd_ShouldAddNewActionToTheRepo(t *testing.T) {

	mongoT.DropDatabase(t)
	want := newActionT("1", "actionA", stateNew, time.Now())
	want.States = []State{want.State}

	err := actionRepo.Add(want)
	require.NoError(t, err)

	var got Action
	mongoT.FindOneT(t, mongo.ActionCollectionId, bson.M{"_id": "1"}, &got)
	assert.Equal(t, want, got)
}

func TestAdd_ShouldAddNewActionIncludingFlowNameToTheRepo(t *testing.T) {

	mongoT.DropDatabase(t)
	state := State{Value: stateNew, Time: time.Now().Round(time.Millisecond)}
	want := Action{
		Id:       "1",
		Name:     "actionA",
		FlowName: "flowA",
		State:    state,
		States:   []State{state},
	}

	err := actionRepo.Add(want)
	require.NoError(t, err)

	var got Action
	mongoT.FindOneT(t, mongo.ActionCollectionId, bson.M{"_id": "1"}, &got)
	assert.Equal(t, want, got)
}

func TestUpdate_ShouldFailToAddActionWithExistingID(t *testing.T) {

	mongoT.DropDatabase(t)
	action := newActionT("1", "actionA", stateNew, time.Now())
	mongoT.Insert(t, mongo.ActionCollectionId, action)

	err := actionRepo.Add(action)

	assert.Error(t, err)
	assert.True(t, mgo.IsDup(err))
}

func TestFindNew_ShouldReturnOldestActionWhichCanBeHandledByPack(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newPackActionT("packA", "1", "actionA", stateNew, time.Now()))
	want := newPackActionT("packA", "2", "actionA", stateNew, time.Now().Add(-1*time.Hour))
	want.States = []State{want.State}
	mongoT.Insert(t, mongo.ActionCollectionId, want)
	mongoT.Insert(t, mongo.ActionCollectionId, newPackActionT("packA", "3", "actionA", statePending, time.Now().Add(-2*time.Hour)))

	got, err := actionRepo.FindNew(Pack{Name: "packA"}, "")
	require.NoError(t, err)

	assert.Equal(t, want, *got)
}

func TestFindNew_ShouldReturnNilWhenThereIsNoNewActionsForAPack(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newPackActionT("packA", "1", "actionA", stateNew, time.Now()))

	got, err := actionRepo.FindNew(Pack{Name: "packWithoutActions"}, "")
	require.NoError(t, err)

	assert.True(t, got == nil)
}

func TestFindNew_ShouldReturnOldestPackNewActionWithMatchingName(t *testing.T) {

	mongoT.DropDatabase(t)
	want := newPackActionT("packA", "1", "actionA", stateNew, time.Now())
	want.States = []State{want.State}
	mongoT.Insert(t, mongo.ActionCollectionId, want)

	got, err := actionRepo.FindNew(Pack{Name: "packA"}, "actionA")
	require.NoError(t, err)

	assert.Equal(t, want, *got)
}

func TestFindNew_ShouldReturnNilWhenThereIsNoPackNewActionMatchingName(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newPackActionT("packA", "1", "actionB", stateNew, time.Now()))

	got, err := actionRepo.FindNew(Pack{Name: "packA"}, "actionA")
	require.NoError(t, err)

	assert.True(t, got == nil)
}

func TestFindNew_ShouldReturnAnyOldestPackNewActionWhenNameIsNotSpecified(t *testing.T) {

	mongoT.DropDatabase(t)
	want := newPackActionT("packA", "1", "actionB", stateNew, time.Now())
	want.States = []State{want.State}
	mongoT.Insert(t, mongo.ActionCollectionId, want)

	got, err := actionRepo.FindNew(Pack{Name: "packA"}, "")
	require.NoError(t, err)

	assert.Equal(t, want, *got)
}

func TestFindNew_ShouldReturnActionWithOldestNewState(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newPackActionT("packA", "1", "actionA", stateNew, time.Now()))
	want := newPackActionT("packA", "2", "actionA", stateNew, time.Now().Add(-1*time.Hour))
	want.States = []State{want.State}
	mongoT.Insert(t, mongo.ActionCollectionId, want)
	mongoT.Insert(t, mongo.ActionCollectionId, newPackActionT("packA", "3", "actionA", stateNew, time.Now()))

	got, err := actionRepo.FindNew(Pack{Name: "packA"}, "actionA")
	require.NoError(t, err)

	assert.Equal(t, want, *got)
}

func TestFindNew_ShouldReturnNilWhenThereIsNoActionWithNewStateMatchingPack(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newPackActionT("packA", "1", "actionA", statePending, time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newPackActionT("packA", "2", "actionA", stateFatal, time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newPackActionT("packA", "3", "actionA", stateSuccess, time.Now()))

	got, err := actionRepo.FindNew(Pack{Name: "packA"}, "actionA")
	require.NoError(t, err)

	assert.True(t, got == nil)
}

func TestFindCorrelated_ShouldReturnAllCorrelatedActions(t *testing.T) {

	mongoT.DropDatabase(t)
	actionA := newActionT("1", "actionA", stateNew, time.Now())
	actionA.CorrelationId = "correlated"
	actionB := newActionT("2", "actionB", statePending, time.Now())
	actionB.CorrelationId = "correlated"
	actionC := newActionT("3", "actionC", stateSuccess, time.Now())
	actionC.CorrelationId = "correlated"
	actionD := newActionT("4", "someUnrelatedAction", stateNew, time.Now())
	actionD.CorrelationId = "IamFromTheDifferentBunch"
	mongoT.Insert(t, mongo.ActionCollectionId, actionA)
	mongoT.Insert(t, mongo.ActionCollectionId, actionB)
	mongoT.Insert(t, mongo.ActionCollectionId, actionC)
	mongoT.Insert(t, mongo.ActionCollectionId, actionD)

	got, err := actionRepo.FindCorrelated("correlated")
	require.NoError(t, err)

	//remove correlationIds, packName & name as we don't select them from mgo
	actionA.CorrelationId = ""
	actionA.PackName = ""
	actionA.Name = ""
	actionB.CorrelationId = ""
	actionB.PackName = ""
	actionB.Name = ""
	actionC.CorrelationId = ""
	actionC.PackName = ""
	actionC.Name = ""
	want := []Action{actionA, actionB, actionC}
	assert.Equal(t, want, got)
}

func TestFindCorrelated_ShouldReturnNilIfThereAreNoCorrelatedActions(t *testing.T) {

	mongoT.DropDatabase(t)
	action := newActionT("1", "actionA", stateNew, time.Now())
	action.CorrelationId = "someCorrelationID"
	mongoT.Insert(t, mongo.ActionCollectionId, action)

	got, err := actionRepo.FindCorrelated("nonExistingCorrelationId")
	require.NoError(t, err)

	assert.True(t, got == nil)
}

func TestGet_ShouldReturnPackActionBySpecifiedIdIfOneExists(t *testing.T) {

	mongoT.DropDatabase(t)
	want := newActionT("matchingActionId", "actionA", stateNew, time.Now())
	want.States = []State{want.State}
	mongoT.Insert(t, mongo.ActionCollectionId, want)

	got, err := actionRepo.Get("matchingActionId")
	require.NoError(t, err)
	require.True(t, got != nil)

	assert.Equal(t, want, *got)
}

func TestGet_ShouldReturnErrorWhenActionIdDoesNotMatch(t *testing.T) {

	mongoT.DropDatabase(t)
	want := newActionT("actionId", "actionA", stateNew, time.Now())
	mongoT.Insert(t, mongo.ActionCollectionId, want)

	_, err := actionRepo.Get("notMatchingActionId")

	assert.EqualError(t, err, ActionNotFoundErr.Error())
}

func TestUpdate_ShouldUpdateActionWhenPreviousStateIsCorrect(t *testing.T) {

	mongoT.DropDatabase(t)
	action := newActionT("1", "actionA", stateNew, time.Now())
	action.States = []State{action.State}
	mongoT.Insert(t, mongo.ActionCollectionId, action)

	action.prevState.Value = stateNew
	action.State.Value = statePending
	action.States = append(action.States, action.State)
	err := actionRepo.Update(action)
	require.NoError(t, err)

	var gotAction Action
	mongoT.FindOneT(t, mongo.ActionCollectionId, bson.M{"_id": "1"}, &gotAction)
	action.prevState = State{}
	assert.Equal(t, action, gotAction)
}

func TestUpdate_ShouldFailToFindAndUpdateActionWhenPreviousStateIsIncorrect(t *testing.T) {

	mongoT.DropDatabase(t)
	action := newActionT("1", "actionA", statePending, time.Now())
	mongoT.Insert(t, mongo.ActionCollectionId, action)

	err := actionRepo.Update(action)
	assert.Error(t, err)
	assert.Equal(t, mgo.ErrNotFound, err)
}

func TestUpdate_ShouldFailToFindAndUpdateActionForIncorrectId(t *testing.T) {

	mongoT.DropDatabase(t)
	action := newActionT("1", "actionA", stateNew, time.Now())
	mongoT.Insert(t, mongo.ActionCollectionId, action)
	actionWithDifferentId := newActionT("differentId", "actionA", stateNew, time.Now())

	err := actionRepo.Update(actionWithDifferentId)
	assert.Error(t, err)
	assert.Equal(t, mgo.ErrNotFound, err)
}

func newActionT(id, name, state string, stateTime time.Time) Action {
	return Action{
		Id:    id,
		Name:  name,
		State: State{Value: state, Time: stateTime.Round(time.Millisecond)},
	}
}

func newPackActionT(packName, id, name, state string, stateTime time.Time) Action {
	return Action{
		Id:       id,
		Name:     name,
		PackName: packName,
		State:    State{Value: state, Time: stateTime.Round(time.Millisecond)},
	}
}
