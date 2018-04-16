// +build acceptance

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

var ExecutionFeatures = []Test{
	{"FlowExecutionTest", FlowExecutionTest},
}

func FlowExecutionTest(t *testing.T) {
	ResetFlyteApi(t)
	packLoc := httpClient.PostResourceAndAssert(flyteApi.PacksURL(), `{"name": "Slack"}`, "pack", t)
	httpClient.PostResourceAndAssert(flyteApi.FlowsURL(), slackFlow, "flow", t)

	// send event
	resp, err := httpClient.Post(packLoc.String()+"/events", `{"event": "MessageSent"}`)
	require.NoError(t, err)
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	time.Sleep(100 * time.Millisecond) // wait for action to be created

	// take action
	resp, err = httpClient.Post(packLoc.String()+"/actions/take", "")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	var action = struct {
		Links []map[string]string
	}{}

	err = json.Unmarshal(b, &action)
	require.NoError(t, err)
	assert.Contains(t, string(b), "SendMessage")

	// finish action
	resp, err = httpClient.Post(action.Links[0]["href"], `{"event": "MessageSent"}`)
	require.NoError(t, err)
	require.Equal(t, http.StatusAccepted, resp.StatusCode)
}

const slackFlow = `
{
    "name": "slack_flow",
    "steps": [
        {
            "id": "slack",
            "event": {
                "packName": "Slack",
                "name": "MessageSent"
            },
            "command": {
                "packName": "Slack",
                "name": "SendMessage"
            }
        }
    ]
}`
