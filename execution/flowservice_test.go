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
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
)

func TestHandleEvent_ShouldTriggerHandleEventForAllCandidateFlows(t *testing.T) {

	//Given
	defer resetFlowRepo()
	flows := []Flow{{UUID: "flowA"}, {UUID: "flowB"}}
	actualEvent := Event{}
	flowRepo = mockFlowRepo{
		findByEvent: func(e Event) ([]Flow, error) {
			actualEvent = e
			return flows, nil
		},
	}

	defer resetFlowEventHandler()
	calledFlowA, calledFlowB := false, false
	var wg sync.WaitGroup
	wg.Add(2)
	flowEventHandler = func(f *Flow, e Event) {
		if e.Name == "MessageSent" {
			switch f.UUID {
			case "flowA":
				defer wg.Done()
				calledFlowA = true
				return
			case "flowB":
				defer wg.Done()
				calledFlowB = true
				return
			}
		}
		t.Fatal("Should not get here")
	}

	//When
	expectedEvent := Event{Name: "MessageSent"}
	flowService{}.HandleEvent(expectedEvent)

	//Then
	//waitWithTimeout(wg, 1000*time.Millisecond)
	wg.Wait()
	assert.True(t, calledFlowA, "Should have called event handler on the flowA")
	assert.True(t, calledFlowB, "Should have called event handler on the flowB")
	assert.Equal(t, expectedEvent, actualEvent)
}

func TestHandleAction_ShouldTriggerHandleEventOnExistingFlowForProvidedAction(t *testing.T) {

	//Given
	defer resetFlowRepo()
	actualAction := Action{}
	flowRepo = mockFlowRepo{
		getByAction: func(a Action) (*Flow, error) {
			actualAction = a
			return &Flow{correlationId: "actionFlow"}, nil
		},
	}

	defer resetFlowEventHandler()
	calledFlow := false
	flowEventHandler = func(f *Flow, e Event) {
		fmt.Println(f.correlationId)
		if e.Name == "ResultEvent" && f.correlationId == "actionFlow" {
			calledFlow = true
			return
		}
		t.Fatal("Should not get here")
	}

	//When
	expectedAction := Action{CorrelationId: "actionFlow", Result: Event{Name: "ResultEvent"}}
	flowService{}.HandleAction(expectedAction)

	//Then
	assert.True(t, calledFlow, "Should have called event handler on the flow")
	assert.Equal(t, expectedAction, actualAction)
}

// --- mocks & helpers ---

type mockFlowRepo struct {
	getByAction func(a Action) (*Flow, error)
	findByEvent func(e Event) ([]Flow, error)
}

func (r mockFlowRepo) GetByAction(a Action) (*Flow, error) {
	return r.getByAction(a)

}
func (r mockFlowRepo) FindByEvent(e Event) ([]Flow, error) {
	return r.findByEvent(e)
}

func resetFlowRepo()         { flowRepo = flowMgoRepo{} }
func resetFlowEventHandler() { flowEventHandler = flowEventHandlerFn }

func waitWithTimeout(wg sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		wg.Wait()
		c <- struct{}{}
	}()
	select {
	case <-c:
		return true
	case <-time.After(timeout):
		return false
	}
}
