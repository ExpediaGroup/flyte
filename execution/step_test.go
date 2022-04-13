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
	"testing"
)

func TestStepExecute_ShouldReturnAction(t *testing.T) {

	step := Step{
		Context: map[string]string{
			"contextEnv": "{{ Context.parentContextEnv}}",
		},
		Event: EventDef{
			Name:     "matchingEventName",
			PackName: "matchingPackName",
			PackLabels: map[string]string{
				"env": "{{ Event.Payload.eventEnv }}",
			},
		},
		Criteria: "{{ Event.Payload.eventEnv == Context.contextEnv }}",
		Command: Command{
			PackName: "packA",
			PackLabels: map[string]string{
				"env": "{{ Event.Payload.eventEnv }}",
			},
			Name: "actionA",
			Input: map[string]interface{}{
				"parentContextEnv": "{{ Context.parentContextEnv }}",
				"contextEnv":       "{{ Context.contextEnv }}",
				"eventEnv":         "{{ Event.Payload.eventEnv }}",
			},
		},
	}

	event := newEventT("matchingEventName", "matchingPackName")
	event.Pack.Labels = map[string]string{"env": "dev"}
	event.Payload = map[string]interface{}{"eventEnv": "dev"}

	got, err := step.Execute(event, map[string]string{"parentContextEnv": "dev"})
	require.NoError(t, err)

	assert.NotNil(t, got)
	assert.NotEmpty(t, got.Id)

	state := State{
		Value: stateNew,
		Time:  got.State.Time,
	}

	want := Action{
		Id:         got.Id,
		Name:       "actionA",
		PackName:   "packA",
		PackLabels: map[string]string{"env": "dev"},
		Input: map[string]interface{}{
			"parentContextEnv": "dev",
			"contextEnv":       "dev",
			"eventEnv":         "dev",
		},
		State: state,
		States: []State{state},

		Context: map[string]string{
			"parentContextEnv": "dev",
			"contextEnv":       "dev",
		},
		Trigger: event,
	}
	assert.Equal(t, want, *got)
}

func TestStepExecute_ShouldReturnErrorWhenThereIsErrorWhileResolvingContext(t *testing.T) {

	step := Step{
		Context: map[string]string{
			"invalidTemplate": "{{invalidTemplate",
		},
	}

	_, err := step.Execute(newEventT("eventA", "packA"), map[string]string{})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "error resolving context")
}

func TestStepExecute_ShouldReturnNilWhenEventNameDoesNotMatch(t *testing.T) {

	step := Step{
		Event: EventDef{
			Name:     "eventA",
			PackName: "matchingPackName",
		},
	}

	action, err := step.Execute(newEventT("nonMatchingEventName", "matchingPackName"), map[string]string{})
	require.NoError(t, err)

	assert.Nil(t, action)
}

func TestStepExecute_ShouldReturnNilWhenEventPackNameDoesNotMatch(t *testing.T) {

	step := Step{
		Event: EventDef{
			Name:     "matchingEventName",
			PackName: "packA",
		},
	}

	action, err := step.Execute(newEventT("matchingEventName", "nonMatchingPackName"), map[string]string{})
	require.NoError(t, err)

	assert.Nil(t, action)
}

func TestStepExecute_ShouldReturnNilWhenEventPackLabelsDoNotMatch(t *testing.T) {

	step := Step{
		Event: EventDef{
			Name:     "matchingEventName",
			PackName: "matchingPackName", PackLabels: map[string]string{"matchingLabelName": "labelValue"},
		},
	}

	event := newEventT("matchingEventName", "matchingPackName")
	event.Pack.Labels = map[string]string{"matchingLabelName": "nonMatchingLabelValue"}
	action, err := step.Execute(event, map[string]string{})
	require.NoError(t, err)

	assert.Nil(t, action)
}

func TestStepExecute_ShouldReturnErrorWhenThereIsErrorWhileResolvingEventPackLabels(t *testing.T) {

	step := Step{
		Event: EventDef{
			Name:     "eventA",
			PackName: "packA",
			PackLabels: map[string]string{
				"invalidLabel": "{{invalidTemplate",
			},
		},
	}

	_, err := step.Execute(newEventT("eventA", "packA"), map[string]string{})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "error resolving pack labels")
}

func TestStepExecute_ShouldReturnNilWhenCriteriaAreNotMet(t *testing.T) {

	step := Step{
		Event: EventDef{
			Name:     "eventA",
			PackName: "packA",
		},
		Criteria: "false",
	}

	action, err := step.Execute(newEventT("eventA", "packA"), map[string]string{})
	require.NoError(t, err)

	assert.Nil(t, action)
}

func TestStepExecute_ShouldReturnErrorWhenThereIsAnErrorResolvingCriteria(t *testing.T) {

	step := Step{
		Event: EventDef{
			Name:     "eventA",
			PackName: "packA",
		},
		Criteria: "{{invalidTemplate",
	}

	_, err := step.Execute(newEventT("eventA", "packA"), map[string]string{})
	require.Error(t, err)

	assert.Contains(t, err.Error(), "error resolving criteria")
}

// --- helpers ---

func newEventT(name, packName string) Event {
	return Event{Name: name, Pack: Pack{Id: packName, Name: packName}}
}
