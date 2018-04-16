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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

var PackFeatures = []Test{
	{"Add Pack", AddPack},
	{"Get Individual Pack", GetPack},
	{"Get Packs", GetPacks},
	{"Delete Pack", DeletePack},
}

func AddPack(t *testing.T) {

	ResetFlyteApi(t)

	resp, err := httpClient.Post(flyteApi.PacksURL(), jiraPackDef)
	require.NoError(t, err)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	loc, err := resp.Location()
	require.NoError(t, err, "Error getting location from response")

	assert.Equal(t, flyteApi.PacksURL()+"/Jira", loc.String())
}

func GetPack(t *testing.T) {

	ResetFlyteApi(t)
	loc := httpClient.PostResourceAndAssert(flyteApi.PacksURL(), jiraPackDef, "pack", t)

	resp, err := httpClient.Get(loc.String())
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var json map[string]interface{}
	unmarshalResponse(t, resp.Body, &json)
	assert.Equal(t, "Jira", json["id"].(string))
}

func GetPacks(t *testing.T) {

	ResetFlyteApi(t)
	bambooPackLoc := httpClient.PostResourceAndAssert(flyteApi.PacksURL(), bambooPackDef, "pack", t)
	jiraPackLoc := httpClient.PostResourceAndAssert(flyteApi.PacksURL(), jiraPackDef, "pack", t)

	resp, err := httpClient.Get(flyteApi.PacksURL())
	require.NoError(t, err)

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, string(body), bambooPackLoc.String())
	assert.Contains(t, string(body), jiraPackLoc.String())
}

func DeletePack(t *testing.T) {
	bambooPackLoc := httpClient.PostResourceAndAssert(flyteApi.PacksURL(), bambooPackDef, "pack", t)

	resp, err := httpClient.Delete(bambooPackLoc.String())
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

const jiraPackDef = `
{
    "name": "Jira",
    "commands": [
        {
            "name": "UpdateTicket",
            "events": ["TicketUpdated"]
        }
    ],
    "events": [
        {
            "name": "TicketCreated"
        },
        {
            "name": "TicketUpdated"
        }
    ]
}`

const bambooPackDef = `
{
    "name": "Bamboo",
    "commands": [
        {
            "name": "CreatePlan",
            "events": ["PlanCreated"]
        }
    ],
    "events": [
        {
            "name": "PlanCreated"
        }
    ]
}`
