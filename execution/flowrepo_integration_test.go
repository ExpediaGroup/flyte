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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2/bson"
	"github.com/HotelsDotCom/flyte/mongo"
	"testing"
)

func TestGetByAction_ShouldReturnFlowWithAllCorrelatedActions(t *testing.T) {

	mongoT.DropDatabase(t)

	want := Flow{
		Name:  "flowA",
		UUID:  "existingFlow",
		Steps: []Step{{Id: "stepA"}, {Id: "stepB"}, {Id: "stepC"}},
	}
	mongoT.Insert(t, mongo.HistoryCollectionId, want)
	actionA := Action{Id: "actionA", FlowName: "flowA", FlowUUID: "existingFlow", CorrelationId: "correlatedActions", StepId: "stepA", Context: map[string]string{"ctxKey": "ctxVal"}, State: State{Value: stateSuccess}}
	actionB := Action{Id: "actionB", FlowName: "flowA", FlowUUID: "existingFlow", CorrelationId: "correlatedActions", StepId: "stepB", Context: map[string]string{"ctxKeyB": "ctxBVal"}, State: State{Value: statePending}}
	mongoT.Insert(t, mongo.ActionCollectionId, actionA)
	mongoT.Insert(t, mongo.ActionCollectionId, actionB)

	got, err := flowRepo.GetByAction(actionA)
	require.NoError(t, err)
	require.NotNil(t, got)

	want.correlationId = actionA.CorrelationId
	want.context = actionA.Context
	actionA.CorrelationId = ""
	actionA.FlowName = ""
	actionA.FlowUUID = ""
	actionA.Context = nil
	actionB.CorrelationId = ""
	actionB.FlowName = ""
	actionB.FlowUUID = ""
	actionB.Context = nil
	want.actions = map[string]Action{"stepA": actionA, "stepB": actionB}
	assert.Equal(t, want, *got)
}

func TestGetByAction_ShouldReturnErrorWhenFlowDoesNotExist(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{UUID: "existingFlow"})

	action := Action{FlowUUID: "nonExistingFlow", CorrelationId: "correlatedActions"}
	_, err := flowRepo.GetByAction(action)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "flow with uuid=nonExistingFlow not found")
}

func TestFindCandidatesByTriggerEvent_ShouldReturnListOfFlowMatchingTriggerEventNameAndPackName(t *testing.T) {

	mongoT.DropDatabase(t)

	flowA := Flow{UUID: "flowA", Steps: []Step{{Event: EventDef{Name: "eventOK", PackName: "packOK"}}}}
	flowB := Flow{UUID: "flowB", Steps: []Step{{Event: EventDef{Name: "eventNotOK", PackName: "packNotOk"}}, {Event: EventDef{Name: "eventOK", PackName: "packOK"}}}}
	mongoT.Insert(t, mongo.FlowCollectionId, flowA)
	mongoT.Insert(t, mongo.FlowCollectionId, flowB)

	got, err := flowRepo.FindByEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})
	require.NoError(t, err)
	require.NotNil(t, got)

	prevCorrelation := ""
	for i, f := range got {
		assert.NotEmpty(t, f.correlationId)
		assert.NotEqual(t, prevCorrelation, f.correlationId)
		prevCorrelation = f.correlationId
		got[i].correlationId = ""
	}

	flowA.actions = map[string]Action{}
	flowA.context = map[string]string{}
	flowB.actions = map[string]Action{}
	flowB.context = map[string]string{}
	want := []Flow{flowA, flowB}
	assert.Equal(t, want, got)
}

func TestFindCandidatesByTriggerEvent_ShouldReturnEmptyListIfThereAreNoCandidateFlows(t *testing.T) {

	mongoT.DropDatabase(t)

	flowA := Flow{UUID: "flowA", Steps: []Step{{Event: EventDef{Name: "eventOK", PackName: "packNotOK"}}}}
	flowB := Flow{UUID: "flowB", Steps: []Step{{Event: EventDef{Name: "eventNotOK", PackName: "packOK"}}}}
	flowC := Flow{UUID: "flowC", Steps: []Step{{DependsOn: []string{"notOK"}, Event: EventDef{Name: "eventOK", PackName: "packOK"}}}}
	mongoT.Insert(t, mongo.FlowCollectionId, flowA)
	mongoT.Insert(t, mongo.FlowCollectionId, flowB)
	mongoT.Insert(t, mongo.FlowCollectionId, flowC)

	got, err := flowRepo.FindByEvent(Event{Name: "eventOK", Pack: Pack{Name: "packOK"}})
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Empty(t, got)
}

var flowPrivateRepo = flowMgoRepo{}

func TestGetFlow_ShouldReturnFlowForGivenUUID(t *testing.T) {

	mongoT.DropDatabase(t)

	want := Flow{
		UUID: bson.NewObjectId().Hex(),
		Steps: []Step{
			{
				Id:        "stepA",
				DependsOn: []string{"stepA"},
				Event:     EventDef{Name: "eventA", PackName: "packA"},
				Context:   map[string]string{"ctxKey": "ctxVal"},
				Criteria:  "true",
				Command:   Command{PackName: "packA", PackLabels: map[string]string{"env": "staging"}, Name: "commandA", Input: "input data"},
			},
		},
	}
	mongoT.Insert(t, mongo.HistoryCollectionId, want)

	got, err := flowPrivateRepo.getFlow(want.UUID)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, want, *got)
}

func TestGetFlow_ShouldReturnNilWhenFlowDoesNotExist(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.HistoryCollectionId, Flow{UUID: "someExistingFlow"})

	got, err := flowPrivateRepo.getFlow("nonExistingFlow")
	require.NoError(t, err)

	assert.Nil(t, got)
}
