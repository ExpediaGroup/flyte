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

package flow

import (
	"github.com/HotelsDotCom/flyte/mongo"
	"github.com/HotelsDotCom/flyte/mongo/mongotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"testing"
)

const ttl = 365 * 24 * 60 * 60

var mongoT *mongotest.MongoT

func TestMain(m *testing.M) {
	os.Exit(runTestsWithMongo(m))
}

func runTestsWithMongo(m *testing.M) int {
	mongoT = mongotest.NewMongoT(mongo.DbName)
	defer mongoT.Teardown()

	mongoT.Start()

	mongo.InitSession(mongoT.GetUrl(), ttl)

	return m.Run()
}

func TestAdd_ShouldInsertTheSameFlowIntoTheLatestAndHistoryCollection(t *testing.T) {

	mongoT.DropDatabase(t)

	expectedFlow := newFlow()
	err := flowRepo.Add(expectedFlow)
	require.NoError(t, err)

	assert.Equal(t, expectedFlow, findFlow(t, mongo.FlowCollectionId, expectedFlow.UUID))
	assert.Equal(t, expectedFlow, findFlow(t, mongo.HistoryCollectionId, expectedFlow.UUID))
}

func TestAdd_ShouldInsertFlowIntoTheHistoryCollectionAndUpdateFlowInTheLatestCollection(t *testing.T) {

	mongoT.DropDatabase(t)
	fv1 := newFlow()
	mongoT.Insert(t, mongo.FlowCollectionId, fv1)
	mongoT.Insert(t, mongo.HistoryCollectionId, fv1)

	fv2 := newFlow()
	err := flowRepo.Add(fv2)
	require.NoError(t, err)

	var latestFlowV1 Flow
	err = mongoT.FindOne(mongo.FlowCollectionId, bson.M{"id": fv1.UUID}, &latestFlowV1)
	assert.True(t, err == mgo.ErrNotFound, "Should have returned ErrNotFound exception for flow version 1 in the latest collection")

	assert.Equal(t, fv2, findFlow(t, mongo.FlowCollectionId, fv2.UUID))
	assert.Equal(t, fv1, findFlow(t, mongo.HistoryCollectionId, fv1.UUID))
	assert.Equal(t, fv2, findFlow(t, mongo.HistoryCollectionId, fv2.UUID))
}

func TestRemove_ShouldRemoveExistingFlowOnlyFromTheLatestCollection(t *testing.T) {

	mongoT.DropDatabase(t)
	f := Flow{Name: "flowA"}
	mongoT.Insert(t, mongo.HistoryCollectionId, f)
	mongoT.Insert(t, mongo.FlowCollectionId, f)

	err := flowRepo.Remove(f.Name)
	require.NoError(t, err)

	assert.Equal(t, 0, mongoT.Count(t, mongo.FlowCollectionId))
	assert.Equal(t, 1, mongoT.Count(t, mongo.HistoryCollectionId))
	assert.Equal(t, f, findFlow(t, mongo.HistoryCollectionId, f.UUID))
}

func TestRemove_ShouldReturnErrorWhenFlowDoesNotExistInTheLatestCollection(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{Name: "onlyInHistory"})

	err := flowRepo.Remove("onlyInHistory")

	assert.EqualError(t, err, FlowNotFoundErr.Error())
}

func TestGet_ShouldGetExistingFlowFromTheLatestCollection(t *testing.T) {

	mongoT.DropDatabase(t)
	wantFlow := newFlow()
	mongoT.Insert(t, mongo.FlowCollectionId, wantFlow)

	gotFlow, err := flowRepo.Get(wantFlow.Name)
	require.NoError(t, err)

	assert.Equal(t, wantFlow, *gotFlow)
}

func TestGet_ShouldReturnNilIfFlowDoesNotExistInTheLatestCollection(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{Name: "nonExistingFlowName"})

	fl, err := flowRepo.Get("nonExistingFlowName")

	assert.True(t, fl == nil)
	assert.EqualError(t, err, FlowNotFoundErr.Error())
}

func TestFindAll_ShouldReturnAllLatestFlows(t *testing.T) {

	mongoT.DropDatabase(t)
	f1 := Flow{Name: "flow1", Description: "Flow description"}
	f2 := Flow{Name: "flow2"}
	f3 := Flow{Name: "flow3"}
	mongoT.Insert(t, mongo.FlowCollectionId, f1)
	mongoT.Insert(t, mongo.FlowCollectionId, f2)
	mongoT.Insert(t, mongo.HistoryCollectionId, f1)
	mongoT.Insert(t, mongo.HistoryCollectionId, f2)
	mongoT.Insert(t, mongo.HistoryCollectionId, f3)

	fls, err := flowRepo.FindAll()
	require.NoError(t, err)

	assert.Len(t, fls, 2)
	assert.Contains(t, fls, f1)
	assert.Contains(t, fls, f2)
}

func TestFindAll_ShouldReturnEmptySliceIfThereAreNoLatestFlows(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{Name: "onlyHistoryFlow"})

	fls, err := flowRepo.FindAll()
	require.NoError(t, err)

	assert.Len(t, fls, 0)
}

func findFlow(t *testing.T, cName, uuid string) Flow {
	var f Flow
	mongoT.FindOneT(t, cName, bson.M{"uuid": uuid}, &f)
	return f
}

func newFlow() Flow {
	return Flow{
		UUID:        bson.NewObjectId().Hex(),
		Name:        "flowName",
		Description: "Flow description",
		Steps: []Step{
			{
				Id:      "stepA",
				Event:   Event{Name: "eventA", PackName: "packA"},
				Command: Command{PackName: "packA", Name: "commandA", Input: "input data"},
			},
			{
				Id: "stepB", DependsOn: []string{"stepA"},
				Event:   Event{Name: "actionEventA", PackName: "packA"},
				Command: Command{PackName: "packB", PackLabels: map[string]string{"env": "staging"}, Name: "commandB"},
			},
		},
	}
}
