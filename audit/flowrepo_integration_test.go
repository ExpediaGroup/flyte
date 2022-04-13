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

package audit

import (
	"github.com/ExpediaGroup/flyte/mongo"
	"github.com/ExpediaGroup/flyte/mongo/mongotest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2/bson"
	"os"
	"testing"
	"time"
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

func TestFindFlows_ShouldReturnFlowsSortedByActionWithLatestStateTimestamp(t *testing.T) {

	mongoT.DropDatabase(t)
	actionA := newActionT("flowA", "", "stepA", time.Now().Add(-2*time.Hour))
	actionA.FlowUUID = "defA"
	actionA.States = []State{actionA.State}
	actionB := newActionT("flowA", "", "stepB", time.Now().Add(-1*time.Hour))
	actionB.FlowUUID = "defA"
	actionB.States = []State{}
	actionC := newActionT("flowB", "", "stepA", time.Now())
	actionC.FlowUUID = "defB"
	actionC.States = []State{}
	actionD := newActionT("flowC", "", "stepA", time.Now().Add(-3*time.Hour))
	actionD.FlowUUID = "defA"
	actionD.States = []State{}
	mongoT.Insert(t, mongo.ActionCollectionId, actionA)
	mongoT.Insert(t, mongo.ActionCollectionId, actionB)
	mongoT.Insert(t, mongo.ActionCollectionId, actionC)
	mongoT.Insert(t, mongo.ActionCollectionId, actionD)
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{UUID: "defA", Steps: []Step{{Id: "stepA"}, {Id: "stepB"}}})
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{UUID: "defB", Steps: []Step{{Id: "stepA"}}})

	got, err := flowRepo.Find(flowsFilter{limit: 50})
	require.NoError(t, err)

	flowA := Flow{UUID: "defA", CorrelationId: "flowA", Steps: []Step{{Id: "stepA"}, {Id: "stepB"}}}
	flowA.Actions = map[string]Action{"stepA": actionA, "stepB": actionB}
	flowB := Flow{UUID: "defB", CorrelationId: "flowB", Steps: []Step{{Id: "stepA"}}}
	flowB.Actions = map[string]Action{"stepA": actionC}
	flowC := Flow{UUID: "defA", CorrelationId: "flowC", Steps: []Step{{Id: "stepA"}, {Id: "stepB"}}}
	flowC.Actions = map[string]Action{"stepA": actionD}
	want := []Flow{flowB, flowA, flowC}
	assert.Equal(t, want, got)
}

func TestFindFlows_ShouldReturnEmptySliceIfThereAreNoFlowsMatchingCriteria(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "someActionA", "stepA", time.Now().Add(-1*time.Hour)))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "someActionB", "stepA", time.Now().Add(-2*time.Hour)))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "someActionC", "stepB", time.Now()))

	filter := flowsFilter{
		actionName: "nonExistingActions",
		limit:      50,
	}
	got, err := flowRepo.Find(filter)
	require.NoError(t, err)

	assert.Empty(t, got)
}

func TestFindCorrelationIds_ShouldReturnDistinctCorrelationIdsSortedByActionWithLatestStateTimestamp(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "", "", time.Now().Add(-1*time.Hour)))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "", "", time.Now().Add(-2*time.Hour)))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "", "", time.Now()))

	got, err := findCorrelationIds(flowsFilter{limit: 50})
	require.NoError(t, err)

	want := []string{"flowB", "flowA"}
	assert.Equal(t, want, got)
}

func TestFindCorrelationIds_ShouldFilterByFlowName(t *testing.T) {

	mongoT.DropDatabase(t)
	actionA := newActionT("flowA", "", "", time.Now())
	actionA.FlowName = "onlyThis"
	actionB := newActionT("flowB", "", "", time.Now())
	actionB.FlowName = "notThis"
	actionC := newActionT("flowC", "", "", time.Now().Add(-1*time.Hour))
	actionC.FlowName = "onlyThis"
	mongoT.Insert(t, mongo.ActionCollectionId, actionA)
	mongoT.Insert(t, mongo.ActionCollectionId, actionB)
	mongoT.Insert(t, mongo.ActionCollectionId, actionC)

	filter := flowsFilter{
		flowName: "onlyThis",
		limit:    50,
	}
	got, err := findCorrelationIds(filter)
	require.NoError(t, err)

	want := []string{"flowA", "flowC"}
	assert.Equal(t, want, got)
}

func TestFindCorrelationIds_ShouldFilterByStepId(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "", "ayePet", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "", "neeWayMan", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowC", "", "ayePet", time.Now().Add(-1*time.Hour)))

	filter := flowsFilter{
		stepId: "ayePet",
		limit:  50,
	}
	got, err := findCorrelationIds(filter)
	require.NoError(t, err)

	want := []string{"flowA", "flowC"}
	assert.Equal(t, want, got)
}

func TestFindCorrelationIds_ShouldFilterByActionName(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "cannyAction", "", time.Now().Add(-1*time.Hour)))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "cannyAction", "", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowC", "notCannyAction", "", time.Now()))

	filter := flowsFilter{
		actionName: "cannyAction",
		limit:      50,
	}
	got, err := findCorrelationIds(filter)
	require.NoError(t, err)

	want := []string{"flowB", "flowA"}
	assert.Equal(t, want, got)
}

func TestFindCorrelationIds_ShouldFilterByActionPackName(t *testing.T) {

	mongoT.DropDatabase(t)
	actionA := newActionT("flowA", "", "", time.Now())
	actionA.PackName = "aah"
	actionB := newActionT("flowB", "", "", time.Now())
	actionB.PackName = "huh"
	actionC := newActionT("flowC", "", "", time.Now().Add(-1*time.Hour))
	actionC.PackName = "huh"
	mongoT.Insert(t, mongo.ActionCollectionId, actionA)
	mongoT.Insert(t, mongo.ActionCollectionId, actionB)
	mongoT.Insert(t, mongo.ActionCollectionId, actionC)

	filter := flowsFilter{
		actionPackName: "huh",
		limit:          50,
	}
	got, err := findCorrelationIds(filter)
	require.NoError(t, err)

	want := []string{"flowB", "flowC"}
	assert.Equal(t, want, got)
}

func TestFindCorrelationIds_ShouldFilterByActionPackLabels(t *testing.T) {

	mongoT.DropDatabase(t)
	labels := map[string]string{"env": "dev", "instance": "1"}
	actionA := newActionT("flowA", "", "", time.Now())
	actionA.PackLabels = labels
	actionB := newActionT("flowB", "", "", time.Now())
	actionC := newActionT("flowC", "", "", time.Now().Add(-1*time.Hour))
	actionC.PackLabels = labels
	mongoT.Insert(t, mongo.ActionCollectionId, actionA)
	mongoT.Insert(t, mongo.ActionCollectionId, actionB)
	mongoT.Insert(t, mongo.ActionCollectionId, actionC)

	filter := flowsFilter{
		actionPackLabels: labels,
		limit:            50,
	}
	got, err := findCorrelationIds(filter)
	require.NoError(t, err)

	want := []string{"flowA", "flowC"}
	assert.Equal(t, want, got)
}

func TestFindCorrelationIds_ShouldSkipFirstNItems(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "cannyAction", "", time.Now().Add(-1*time.Hour)))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "cannyAction", "", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowC", "notCannyAction", "", time.Now()))

	filter := flowsFilter{
		actionName: "cannyAction",
		skip:       1,
		limit:      50,
	}
	got, err := findCorrelationIds(filter)
	require.NoError(t, err)

	want := []string{"flowA"}
	assert.Equal(t, want, got)
}

func TestFindCorrelationIds_ShouldReturnNumberOfItemsSpecifiedByLimit(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "", "", time.Now().Add(-2*time.Hour)))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "", "", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowC", "", "", time.Now().Add(-1*time.Hour)))

	filter := flowsFilter{
		limit: 1,
	}
	got, err := findCorrelationIds(filter)
	require.NoError(t, err)

	want := []string{"flowB"}
	assert.Equal(t, want, got)
}

func TestFindCorrelationIds_ShouldReturnEmptySliceIfThereAreNoMatchingActions(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "neeWayMan", "", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "neeWayMan", "", time.Now()))

	filter := flowsFilter{
		actionName: "ayePet",
		limit:      50,
	}
	got, err := findCorrelationIds(filter)
	require.NoError(t, err)

	assert.Empty(t, got)
}

func TestFindActionsByCorrelationIds_ShouldReturnAllMatchingActions(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "", "", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "", "", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowC", "", "", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "", "", time.Now()))

	got, err := findActionsByCorrelationIds([]string{"flowA", "flowB"})
	require.NoError(t, err)

	assert.Len(t, got, 3)
}

func TestFindActionsByCorrelationIds_ShouldReturnEmptySliceIfThereAreNoMatchingActions(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowA", "", "", time.Now()))
	mongoT.Insert(t, mongo.ActionCollectionId, newActionT("flowB", "", "", time.Now()))

	got, err := findActionsByCorrelationIds([]string{"flowC", "flowD"})
	require.NoError(t, err)

	assert.Empty(t, got)
}

func TestGetFlow_ShouldReturnFlowByUUID(t *testing.T) {

	mongoT.DropDatabase(t)
	want := Flow{
		UUID:  "expectedFlow",
		Steps: []Step{{Id: "stepA"}, {Id: "stepB"}, {Id: "stepC"}},
	}
	mongoT.Insert(t, mongo.HistoryCollectionId, want)
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{UUID: "someOtherFlow"})

	got, err := getFlow("expectedFlow")
	require.NoError(t, err)

	assert.Equal(t, want, *got)
}

func TestGetFlow_ShouldReturnErrorWhenFlowDoesNotExist(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{UUID: "someOtherFlow"})

	_, err := getFlow("nonExistingFlow")

	assert.EqualError(t, err, "not found")
}

func TestGetFlow_ShouldReturnFlowWithAllCorrelatedActions(t *testing.T) {

	mongoT.DropDatabase(t)
	actionA := newActionT("flowA", "", "stepA", time.Now().Add(-2*time.Hour))
	actionA.FlowUUID = "defA"
	actionA.States = []State{}
	actionB := newActionT("flowA", "", "stepB", time.Now().Add(-1*time.Hour))
	actionB.FlowUUID = "defA"
	actionB.States = []State{}
	mongoT.Insert(t, mongo.ActionCollectionId, actionA)
	mongoT.Insert(t, mongo.ActionCollectionId, actionB)
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{UUID: "defA", Steps: []Step{{Id: "stepA"}, {Id: "stepB"}}})

	got, err := flowRepo.Get("flowA")
	require.NoError(t, err)

	want := Flow{UUID: "defA", CorrelationId: "flowA", Steps: []Step{{Id: "stepA"}, {Id: "stepB"}}}
	want.Actions = map[string]Action{"stepA": actionA, "stepB": actionB}
	assert.Equal(t, want, *got)
}

// --- helpers ---

func newActionT(correlationId, actionName, stepId string, stateTime time.Time) Action {
	return Action{
		Id:            bson.NewObjectId().Hex(),
		CorrelationId: correlationId,
		Name:          actionName,
		StepId:        stepId,
		State:         State{Time: stateTime.Round(time.Millisecond)},
	}
}
