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
	"errors"
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFlows_ShouldReturnListOfFlowsWithLinks_WhenFlowsExist(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		find: func(filter flowsFilter) ([]Flow, error) {
			flowA := Flow{
				Name:          "flowDefA",
				UUID:          "flowDefAV1",
				CorrelationId: "flowA",
				Steps:         []Step{{Id: "stepA", Event: EventDef{Name: "eventA", PackName: "packA"}}},
				Actions:       map[string]Action{"stepA": {Name: filter.actionName, StepId: "stepA", States: []State{{}}}},
			}
			flowB := Flow{
				Name:          "flowDefA",
				UUID:          "flowDefAV2",
				CorrelationId: "flowB",
				Steps:         []Step{{Id: "stepA", Event: EventDef{Name: "eventA", PackName: "packA"}}},
				Actions:       map[string]Action{"stepA": {Name: filter.actionName, StepId: "stepA"}},
			}

			flows := []Flow{flowA, flowB}
			return flows, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/flows?actionName=actionA", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetFlows(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	expectedBody := `{"flows":[{"name":"flowDefA","uuid":"flowDefAV1","correlationId":"flowA","steps":[{"id":"stepA","event":{"name":"eventA","packName":"packA"},"command":{"name":"","packName":"","input":null}}],"actions":{"stepA":{"id":"","name":"actionA","packName":"","state":{"value":"","time":"0001-01-01T00:00:00Z"},"states":[{"value":"","time":"0001-01-01T00:00:00Z"}],"correlationId":"","flowName":"","flowUUID":"","stepId":"stepA","trigger":{"event":"","pack":{"id":"","name":""},"createdAt":"0001-01-01T00:00:00Z","receivedAt":"0001-01-01T00:00:00Z"},"result":{"event":"","pack":{"id":"","name":""},"createdAt":"0001-01-01T00:00:00Z","receivedAt":"0001-01-01T00:00:00Z"}}},"links":[{"href":"http://example.com/v1/audit/flows/flowA","rel":"self"}]},{"name":"flowDefA","uuid":"flowDefAV2","correlationId":"flowB","steps":[{"id":"stepA","event":{"name":"eventA","packName":"packA"},"command":{"name":"","packName":"","input":null}}],"actions":{"stepA":{"id":"","name":"actionA","packName":"","state":{"value":"","time":"0001-01-01T00:00:00Z"},"correlationId":"","flowName":"","flowUUID":"","stepId":"stepA","trigger":{"event":"","pack":{"id":"","name":""},"createdAt":"0001-01-01T00:00:00Z","receivedAt":"0001-01-01T00:00:00Z"},"result":{"event":"","pack":{"id":"","name":""},"createdAt":"0001-01-01T00:00:00Z","receivedAt":"0001-01-01T00:00:00Z"}}},"links":[{"href":"http://example.com/v1/audit/flows/flowB","rel":"self"}]}],"links":[{"href":"http://example.com/v1/audit/flows","rel":"self"},{"href":"http://example.com/v1","rel":"up"},{"href":"http://example.com/swagger#/flowExecs","rel":"help"}]}`
	assert.JSONEq(t, expectedBody, string(body))
}

func TestGetFlows_ShouldReturnListOfFlowsFilteredByFlowName(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		find: func(filter flowsFilter) ([]Flow, error) {
			require.Equal(t, "flowA", filter.flowName)
			flowA := Flow{
				Name:          "flowA",
				UUID:          "flowDefA",
				CorrelationId: "flowA",
				Steps:         []Step{{Id: "stepA", Event: EventDef{Name: "eventA", PackName: "packA"}}},
				Actions:       map[string]Action{"stepA": {Name: "actionA", StepId: "stepA"}},
			}

			flows := []Flow{flowA}
			return flows, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/flows?flowName=flowA", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetFlows(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	expectedBody := `{"flows":[{"name":"flowA","uuid":"flowDefA","correlationId":"flowA","steps":[{"id":"stepA","event":{"name":"eventA","packName":"packA"},"command":{"name":"","packName":"","input":null}}],"actions":{"stepA":{"id":"","name":"actionA","packName":"","state":{"value":"","time":"0001-01-01T00:00:00Z"},"correlationId":"","flowName":"","flowUUID":"","stepId":"stepA","trigger":{"event":"","pack":{"id":"","name":""},"createdAt":"0001-01-01T00:00:00Z","receivedAt":"0001-01-01T00:00:00Z"},"result":{"event":"","pack":{"id":"","name":""},"createdAt":"0001-01-01T00:00:00Z","receivedAt":"0001-01-01T00:00:00Z"}}},"links":[{"href":"http://example.com/v1/audit/flows/flowA","rel":"self"}]}],"links":[{"href":"http://example.com/v1/audit/flows","rel":"self"},{"href":"http://example.com/v1","rel":"up"},{"href":"http://example.com/swagger#/flowExecs","rel":"help"}]}`
	assert.Equal(t, expectedBody, string(body))
}

func TestGetFlows_ShouldExtractFilterParametersFromTheRequest(t *testing.T) {

	defer resetFlowRepo()
	var got flowsFilter
	flowRepo = mockFlowRepo{
		find: func(filter flowsFilter) ([]Flow, error) {
			got = filter
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/flows?flowName=flowA&stepId=stepA&actionName=actionA&actionPackName=packA&actionPackLabels=env:dev,foo:bar&start=10&limit=10", nil)
	w := httptest.NewRecorder()
	GetFlows(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	want := flowsFilter{
		flowName:         "flowA",
		stepId:           "stepA",
		actionName:       "actionA",
		actionPackName:   "packA",
		actionPackLabels: map[string]string{"env": "dev", "foo": "bar"},
		skip:             10,
		limit:            10,
	}
	assert.Equal(t, want, got)
}

func TestGetFlows_ShouldReturnZeroFlows_WhenThereAreNoFlows(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		find: func(filter flowsFilter) ([]Flow, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/flows", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetFlows(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	expectedBody := `{"flows":[],"links":[{"href":"http://example.com/v1/audit/flows","rel":"self"},{"href":"http://example.com/v1","rel":"up"},{"href":"http://example.com/swagger#/flowExecs","rel":"help"}]}`
	assert.Equal(t, expectedBody, string(body))
}

func TestGetFlows_ShouldReturn500_WhenErrorHappens(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		find: func(filter flowsFilter) ([]Flow, error) {
			return nil, errors.New("expected error")
		},
	}

	r := httptest.NewRequest(http.MethodGet, "/v1/audit/flows", nil)
	w := httptest.NewRecorder()
	GetFlows(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "expected error", logMessages[0].Message)
}

func TestGetFlow_ShouldReturnFlowWithActions(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		get: func(correlationId string) (*Flow, error) {
			flow := &Flow{
				Name:          "flowDef",
				UUID:          "flowDefV1",
				CorrelationId: correlationId,
				Steps:         []Step{{Id: "stepA", Event: EventDef{Name: "eventA", PackName: "packA"}}},
				Actions:       map[string]Action{"stepA": {StepId: "stepA", CorrelationId: correlationId}},
			}
			return flow, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/flows/existingFlow?:flowName=existingFlow", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetFlow(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	expectedBody := `{"name":"flowDef","uuid":"flowDefV1","correlationId":"","steps":[{"id":"stepA","event":{"name":"eventA","packName":"packA"},"command":{"name":"","packName":"","input":null}}],"actions":{"stepA":{"id":"","name":"","packName":"","state":{"value":"","time":"0001-01-01T00:00:00Z"},"correlationId":"","flowName":"","flowUUID":"","stepId":"stepA","trigger":{"event":"","pack":{"id":"","name":""},"createdAt":"0001-01-01T00:00:00Z","receivedAt":"0001-01-01T00:00:00Z"},"result":{"event":"","pack":{"id":"","name":""},"createdAt":"0001-01-01T00:00:00Z","receivedAt":"0001-01-01T00:00:00Z"}}},"links":[{"href":"http://example.com/v1/audit/flows","rel":"self"}]}`
	assert.Equal(t, expectedBody, string(body))
}

func TestGetFlow_ShouldReturn404ForNonExistingFlow(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		get: func(correlationId string) (*Flow, error) {
			return nil, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/audit/flows/nonExistingFlow", nil)
	w := httptest.NewRecorder()
	GetFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetFlow_ShouldReturn500_WhenErrorHappens(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		get: func(correlationId string) (*Flow, error) {
			return nil, errors.New("expected error")
		},
	}

	r := httptest.NewRequest(http.MethodGet, "/v1/audit/flows/errorFlow", nil)
	w := httptest.NewRecorder()
	GetFlow(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Error finding flow correlationId=: expected error", logMessages[0].Message)
}

// --- mocks & helpers ---

type mockFlowRepo struct {
	get  func(correlationId string) (*Flow, error)
	find func(filter flowsFilter) ([]Flow, error)
}

func (r mockFlowRepo) Get(correlationId string) (*Flow, error) {
	return r.get(correlationId)
}

func (r mockFlowRepo) Find(filter flowsFilter) ([]Flow, error) {
	return r.find(filter)
}

func resetFlowRepo() { flowRepo = flowMgoRepo{} }
