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

func TestAuditAdd_ShouldAddNewActionToTheRepo(t *testing.T) {

	mongoT.DropDatabase(t)
	state := State{Value: stateNew, Time: time.Now().Round(time.Millisecond)}
	want := Action{
		Id:       "1",
		Name:     "actionA",
		FlowName: "flowA",
		State:    state,
		States:   []State{state},
	}

	err := auditRepo.Add(want)
	require.NoError(t, err)

	var got Action
	mongoT.FindOneT(t, mongo.AuditCollectionId, bson.M{"_id": "1"}, &got)
	assert.Equal(t, want, got)
}

func TestAuditAdd_ShouldFailToAddActionWithExistingID(t *testing.T) {

	mongoT.DropDatabase(t)
	action := newActionT("1", "actionA", stateNew, time.Now())
	mongoT.Insert(t, mongo.AuditCollectionId, action)

	err := auditRepo.Add(action)

	assert.Error(t, err)
	assert.True(t, mgo.IsDup(err))
}

func TestAuditUpdate_ShouldUpdateActionWhenPreviousStateIsCorrect(t *testing.T) {

	mongoT.DropDatabase(t)
	action := newActionT("1", "actionA", stateNew, time.Now())
	action.States = []State{action.State}
	mongoT.Insert(t, mongo.AuditCollectionId, action)

	action.prevState.Value = stateNew
	action.State.Value = statePending
	action.States = append(action.States, action.State)
	err := auditRepo.Update(action)
	require.NoError(t, err)

	var gotAction Action
	mongoT.FindOneT(t, mongo.AuditCollectionId, bson.M{"_id": "1"}, &gotAction)
	action.prevState = State{}
	assert.Equal(t, action, gotAction)
}

func TestAuditUpdate_ShouldFailToFindAndUpdateActionWhenPreviousStateIsIncorrect(t *testing.T) {

	mongoT.DropDatabase(t)
	action := newActionT("1", "actionA", statePending, time.Now())
	mongoT.Insert(t, mongo.AuditCollectionId, action)

	err := auditRepo.Update(action)
	assert.Error(t, err)
	assert.Equal(t, mgo.ErrNotFound, err)
}

func TestAuditUpdate_ShouldFailToFindAndUpdateActionForIncorrectId(t *testing.T) {

	mongoT.DropDatabase(t)
	action := newActionT("1", "actionA", stateNew, time.Now())
	mongoT.Insert(t, mongo.AuditCollectionId, action)
	actionWithDifferentId := newActionT("differentId", "actionA", stateNew, time.Now())

	err := auditRepo.Update(actionWithDifferentId)
	assert.Error(t, err)
	assert.Equal(t, mgo.ErrNotFound, err)
}
