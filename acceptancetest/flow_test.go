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
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
)

var FlowFeatures = []Test{
	{"PostingFlow_ShouldReturn201AndLocation", PostingFlow_ShouldReturn201AndLocation},
	{"GetFlows_ShouldReturnAllFlows", GetFlows_ShouldReturnAllFlows},
	{"GetFlow_ShouldReturnSpecifiedFlow", GetFlow_ShouldReturnSpecifiedFlow},
	{"ShouldDeleteSpecifiedFlow", ShouldDeleteSpecifiedFlow},
}

func PostingFlow_ShouldReturn201AndLocation(t *testing.T) {

	ResetFlyteApi(t)

	// post flow def
	resp := httpClient.PostAndAssert(flyteApi.FlowsURL(), pobFlow1, "flow", t)

	// should return 201 and location
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	loc, err := resp.Location()
	require.NoError(t, err)

	path := strings.Split(loc.Path, "/")
	assert.Equal(t, "pob_flow1", path[len(path)-1])
}

func GetFlows_ShouldReturnAllFlows(t *testing.T) {

	ResetFlyteApi(t)

	// post some flows
	httpClient.PostResourceAndAssert(flyteApi.FlowsURL(), pobFlow1, "flow", t)
	httpClient.PostResourceAndAssert(flyteApi.FlowsURL(), pobFlow2, "flow", t)

	// get all the flows
	resp, err := httpClient.Get(flyteApi.FlowsURL())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	flows := map[string][]map[string]interface{}{}
	unmarshalResponse(t, resp.Body, &flows)
	defer resp.Body.Close()

	require.Equal(t, 2, len(flows["flows"]))
	assert.Equal(t, "pob_flow1", flows["flows"][0]["name"].(string))
	assert.Equal(t, "pob_flow2", flows["flows"][1]["name"].(string))
}

func GetFlow_ShouldReturnSpecifiedFlow(t *testing.T) {

	ResetFlyteApi(t)

	// post some flows
	httpClient.PostResourceAndAssert(flyteApi.FlowsURL(), pobFlow1, "flow", t)
	pobFlow2Loc := httpClient.PostResourceAndAssert(flyteApi.FlowsURL(), pobFlow2, "flow", t)

	// get 'pob_flow2' flow
	resp, err := httpClient.Get(pobFlow2Loc.String())
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	flow := map[string]interface{}{}
	unmarshalResponse(t, resp.Body, &flow)
	defer resp.Body.Close()

	assert.Equal(t, "pob_flow2", flow["name"].(string))
}

func ShouldDeleteSpecifiedFlow(t *testing.T) {

	ResetFlyteApi(t)

	// post some flows
	pobFlow1Loc := httpClient.PostResourceAndAssert(flyteApi.FlowsURL(), pobFlow1, "flow", t)
	httpClient.PostResourceAndAssert(flyteApi.FlowsURL(), pobFlow2, "flow", t)

	// verify that they both flows are persisted
	resp, err := httpClient.Get(flyteApi.FlowsURL())
	require.NoError(t, err)

	flows := map[string][]map[string]interface{}{}
	unmarshalResponse(t, resp.Body, &flows)
	defer resp.Body.Close()
	require.Equal(t, 2, len(flows["flows"]))

	// delete 'pob_flow1' flow
	resp, err = httpClient.Delete(pobFlow1Loc.String())
	require.NoError(t, err)

	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// verify that only one flow is persisted
	resp, err = httpClient.Get(flyteApi.FlowsURL())
	require.NoError(t, err)

	flows = map[string][]map[string]interface{}{}
	unmarshalResponse(t, resp.Body, &flows)
	defer resp.Body.Close()
	require.Equal(t, 1, len(flows["flows"]))
}

func unmarshalResponse(t *testing.T, body io.Reader, v interface{}) {

	b, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	err = json.Unmarshal(b, v)
	require.NoError(t, err)
}

const pobFlow1 = `
{
    "name": "pob_flow1",
    "steps": []
}`

const pobFlow2 = `
{
    "name": "pob_flow2",
    "steps": []
}`
