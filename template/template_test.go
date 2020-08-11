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

package template

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResolveDoesNotMutateOriginalValue(t *testing.T) {

	originalValue := map[string]interface{}{
		"bool":   true,
		"number": 34,
		"array":  []interface{}{"{{ replace }}", "haha"},
	}

	got, err := Resolve(originalValue, Context{"replace": "updated"})
	require.NoError(t, err)

	want := map[string]interface{}{
		"bool":   true,
		"number": 34,
		"array":  []interface{}{"updated", "haha"},
	}

	assert.Equal(t, want, got)
	assert.Equal(t, []interface{}{"{{ replace }}", "haha"}, originalValue["array"], "Original value should not have been mutated")
}

func TestResolveShouldFailForInterfacesContainingStruct(t *testing.T) {
	structValue := struct{ Name string }{Name: "{{ 'should fail for a struct' }}"}
	_, err := Resolve(structValue, nil)
	assert.EqualError(t, err, "error while evaluating expression: '{Name:{{ 'should fail for a struct' }}}': unsupported kind: struct")
}

func TestResolveShouldFailForInterfacesContainingPointers(t *testing.T) {
	pointerValue := &struct{ Name string }{Name: "{{ 'should fail for a pointer' }}"}
	_, err := Resolve(pointerValue, nil)
	assert.EqualError(t, err, "error while evaluating expression: '&{Name:{{ 'should fail for a pointer' }}}': unsupported kind: ptr")
}

func TestResolveShouldResolveValidPongoTemplates(t *testing.T) {

	var cases = []struct {
		template string
		expected string
	}{
		{
			`some plain text`,
			`some plain text`,
		},
		{
			`Build {{Event.Payload.status}}, first commit was {{Event.Payload.commits.0.commitId}}`,
			`Build success, first commit was 1234`,
		},
		{
			`Build was a success: {{ Event.Payload.statusDetail|key:Event.Payload.status }}`,
			`Build was a success: True`,
		},
		{
			`{{Event.Payload.non.existent.field}}`,
			``,
		},
		{
			`{{ Context.hipchat_room }}`,
			`123`,
		},
		{
			`{{ Event.Payload.status }} and {{ Context.status }}`,
			`success and pending`,
		},
		{
			`{{ "flyte"|upper }}`,
			`FLYTE`,
		},
		{
			`{% if Event.Payload.metadata.rb == "1" %}{{ Event.Payload.metadata._PreviousLabel }}{% else %}{{ Event.Payload.label }}{% endif %}`,
			`CSVC.59.0.29`,
		},
		{
			`{% for commit in Event.Payload.commits %}-{{ commit.commitId }}{% endfor %}`,
			`-1234-2345-3456`,
		},
		{
			`{% set values = Event.Payload.keyValuePairs|kvp %}{{ values.ENV }} {{ values.CMP }} {{ values.LBL }}`,
			`staging flyte 1.2.3`,
		},
		{
			`{{ Event.Payload.keyValuePairs|kvp|key:"bn" }}`,
			`100`,
		},
		{
			`{{ Event.Payload.label|match:"^CSVC.*$" }}`,
			`True`,
		},
		{
			`{{ "pongo version new"|split:" "|index:1 }}`,
			`version`,
		},
		{
			`{{ "pongo version new"|split:" "|last }}`,
			`new`,
		},
		{
			`{{ "pongo version new"|split:' '|first }}`,
			`pongo`,
		},
		{
			`{{ "pongo version new"|extractMatch:'\\w+ \\w+ (\\w+)' }}`,
			`new`,
		},
		{
			`{{ "the good, the bad, and the-ugly"|extractMatch:'.* ([-a-zA-Z0-9_]+)' }}`,
			`the-ugly`,
		},
		{
			`{{ "How do you defeat a Quylthulg?"|extractMatch:'.* ([-a-zA-Z0-9_]+)\\W*' }}`,
			`Quylthulg`,
		},
		{
			`{{ "What is your name?"|extractMatch:'.* (\\w+)$' }}`,
			`What is your name?`,
		},
		{
			`Event Name {{Event.Name}}, Pack name {{Event.Pack.Name}}, Pack label {{Event.Pack.Labels.network}}`,
			`Event Name InventoryUpdateSuccess, Pack name Flyte, Pack label lab`,
		},
		{
			`{{Event.Name == 'InventoryUpdateSuccess'}}`,
			`True`,
		},
		{
			`{{ Context|key:"status" }}`,
			`pending`,
		},
		{
			`{{ base64Encode(Event.Payload.toBase64) }}`,
			`ypjigL_KmA`,
		},
		{
			`{{ base64Decode(Event.Payload.fromBase64) }}`,
			`ʘ‿ʘ`,
		},
		{
			`{{ Event.Payload.timestamp | matchesCron: "18 23 * * *" }}`,
			`True`,
		},
	}

	for i, c := range cases {
		resolved, err := Resolve(c.template, testContext())
		require.NoError(t, err, fmt.Sprintf("case %d: template %s", i, c.template))

		assert.Equal(t, c.expected, resolved, fmt.Sprintf("case %d: template %s", i, c.template))
	}
}

func TestResolveShouldReturnErrorForInvalidTemplate(t *testing.T) {
	invalidTemplate := `{{ nonExistingFunction| }}`
	_, err := Resolve(invalidTemplate, nil)
	assert.Error(t, err)
}

func TestResolveShouldReturnErrorForTemplateWithBackticks(t *testing.T) {
	backtickTemplate := `{% if ` + "`a`" + ` == "a" %}Should fail{% endif %}`
	_, err := Resolve(backtickTemplate, nil)
	assert.Error(t, err)
}

func TestResolveShouldReturnErrorWhenTemplateExecutionPanics(t *testing.T) {
	panicFunc := func() string {
		panic("Upss!!! Something went wrong.")
		return "You will never get this! La la la la la!"
	}
	_, err := Resolve("{{ panicFunc }}", Context{"panicFunc": panicFunc})
	assert.EqualError(t, err, "error while evaluating expression: '{{ panicFunc }}': Upss!!! Something went wrong.")
}

func testContext() Context {
	event := struct {
		Pack    interface{}
		Name    string
		Payload interface{}
	}{
		Pack: struct {
			Name   string
			Labels map[string]string
		}{
			Name:   "Flyte",
			Labels: map[string]string{"env": "staging", "network": "lab"},
		},
		Name: "InventoryUpdateSuccess",
		Payload: map[string]interface{}{
			"status":     "success",
			"product":    "flyte",
			"toBase64":   "ʘ‿ʘ",
			"fromBase64": "ypjigL_KmA",
			"commits":    []struct{ commitId int }{{commitId: 1234}, {commitId: 2345}, {commitId: 3456}},
			"statusDetail": map[string]interface{}{
				"success": true,
			},
			"label": "CSVC.59.0.28",
			"metadata": map[string]interface{}{
				"rb":             "1",
				"_PreviousLabel": "CSVC.59.0.29",
			},
			"keyValuePairs": "bn=100,ENV=staging,CMP=flyte, LBL= 1.2.3",
			"timestamp":     "2018-02-14T23:18:09.0481031Z",
		},
	}
	return Context{
		"Event": event,
		"Context": map[string]interface{}{
			"hipchat_room": "123",
			"status":       "pending",
		},
	}
}
