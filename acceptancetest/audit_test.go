// +build acceptance

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

package acceptancetest

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

var AuditFeatures = []Test{
	{"ShouldFindFlowsWithLinksToIndividualFlows", ShouldFindFlowsWithLinksToIndividualFlows},
}

func ShouldFindFlowsWithLinksToIndividualFlows(t *testing.T) {

	//GIVEN
	ResetFlyteApi(t)
	packLoc := httpClient.PostResourceAndAssert(flyteApi.PacksURL(), `{"name": "Slack"}`, "pack", t)
	flowDef := `{"name": "slack_flow","steps": [{"id": "slack","event": {"packName": "Slack","name": "MessageSent"},"command": {"packName": "Slack","name": "SendMessage"}}]}`
	httpClient.PostResourceAndAssert(flyteApi.FlowsURL(), flowDef, "flow", t)

	// send event
	resp, err := httpClient.Post(packLoc.String()+"/events", `{"event": "MessageSent"}`)
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	time.Sleep(500 * time.Millisecond) // wait for action to be created

	//WHEN
	// find flows
	resp, err = httpClient.Get(flyteApi.FlowExecutionsURL())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	//THEN
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	var response = struct {
		Flows []struct {
			CorrelationId string
		}
	}{}
	require.NoError(t, json.Unmarshal(b, &response))
	require.Len(t, response.Flows, 1)

	// get flow
	resp, err = httpClient.Get(flyteApi.FlowExecutionsURL() + "/" + response.Flows[0].CorrelationId)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	b, err = ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	var flow = struct {
		Name          string
		UUID          string
		CorrelationId string
		Steps         []interface{}
		Actions       map[string]interface{}
	}{}
	require.NoError(t, json.Unmarshal(b, &flow))

	assert.Equal(t, "slack_flow", flow.Name)
	assert.NotEmpty(t, flow.UUID)
	assert.Len(t, flow.Steps, 1)
	assert.Len(t, flow.Actions, 1)
}
