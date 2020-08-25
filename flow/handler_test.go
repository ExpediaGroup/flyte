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

package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestPostFlow_ShouldAddFlowToRepoForValidRequest(t *testing.T) {

	defer resetFlowRepo()
	var actualFlow Flow
	flowRepo = mockFlowRepo{
		add: func(flow Flow) error {
			actualFlow = flow
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/flows", strings.NewReader(redeployFlow))
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	PostFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	location, err := resp.Location()
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/v1/flows/redeploy_flow", location.String())
	var expectedFlow Flow
	err = json.Unmarshal([]byte(redeployFlow), &expectedFlow)
	require.NoError(t, err)
	assert.Equal(t, expectedFlow, actualFlow)
}

func TestPostFlow_TestValidJsonWithMissingField(t *testing.T) {
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	req := httptest.NewRequest(http.MethodPost, "/v1/flows", strings.NewReader(validJsonWithMissingField))
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	PostFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Contains(t, "Cannot convert request to flow: (root): name is required", logMessages[0].Message)

}

func TestPostFlow_ShouldReturn500WhenFlowIsEmpty(t *testing.T) {
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	req := httptest.NewRequest(http.MethodPost, "/v1/flows", strings.NewReader(`{}`))
	w := httptest.NewRecorder()
	PostFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot convert request to flow: (root): name is required", logMessages[0].Message)
}

func TestPostFlow_ShouldReturn500ForInvalidRequest(t *testing.T) {
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	req := httptest.NewRequest(http.MethodPost, "/v1/flows", strings.NewReader("{ 'this is invalid json'"))
	w := httptest.NewRecorder()
	PostFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot convert request to flow: invalid character '\\'' looking for beginning of object key string", logMessages[0].Message)
}

func TestPostFlow_ShouldReturn500_WhenErrorHappens(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		add: func(flow Flow) error {
			return errors.New("something went wrong")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/flows", strings.NewReader(redeployFlow))
	w := httptest.NewRecorder()
	PostFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot add flow to repo flowName=redeploy_flow: something went wrong", logMessages[0].Message)
}

func TestPostFlow_ShouldErrorIfUnableToSetAbsolutePath(t *testing.T) {
	preFileAs := fileAs
	defer func() { fileAs = preFileAs }()
	fileAs = func(path string) (s string, e error) {
		return "", errors.New("unable to generate absolute path")
	}

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	req := httptest.NewRequest(http.MethodPost, "/v1/flows", strings.NewReader(redeployFlow))
	w := httptest.NewRecorder()
	PostFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot convert request to flow: unable to generate absolute path", logMessages[0].Message)
}

func TestPostFlow_ShouldErrorIfUnableToFindFile(t *testing.T) {
	preValidate := validate
	defer func() { validate = preValidate }()
	validate = func(ls gojsonschema.JSONLoader, ld gojsonschema.JSONLoader) (result *gojsonschema.Result, e error) {
		return nil, &os.PathError{Err: fmt.Errorf("file not found")}
	}

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	req := httptest.NewRequest(http.MethodPost, "/v1/flows", strings.NewReader(redeployFlow))
	w := httptest.NewRecorder()
	PostFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Contains(t, logMessages[0].Message, "Cannot convert request to flow: file not found ")
}

func TestGetFlows_ShouldReturnListOfFlowsWithLinks_WhenFlowsExist(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		findAll: func() ([]Flow, error) {
			flows := []Flow{{Name: "flowA"}, {Name: "flowB"}, {Name: "flowC"}}
			return flows, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/flows", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetFlows(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	expectedBody := `{"flows":[{"name":"flowA","links":[{"href":"http://example.com/v1/flows/flowA","rel":"self"}]},{"name":"flowB","links":[{"href":"http://example.com/v1/flows/flowB","rel":"self"}]},{"name":"flowC","links":[{"href":"http://example.com/v1/flows/flowC","rel":"self"}]}],"links":[{"href":"http://example.com/v1/flows","rel":"self"},{"href":"http://example.com/v1","rel":"up"},{"href":"http://example.com/swagger#/flow","rel":"help"}]}`
	assert.Equal(t, expectedBody, string(body))
}

func TestGetFlows_ShouldReturnZeroFlows_WhenThereAreNoFlows(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		findAll: func() ([]Flow, error) {
			return []Flow{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/flows", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetFlows(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	expectedBody := `{"flows":[],"links":[{"href":"http://example.com/v1/flows","rel":"self"},{"href":"http://example.com/v1","rel":"up"},{"href":"http://example.com/swagger#/flow","rel":"help"}]}`
	assert.Equal(t, expectedBody, string(body))
}

func TestGetFlows_ShouldReturn500_WhenErrorHappens(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		findAll: func() ([]Flow, error) {
			return nil, errors.New("something went wrong")
		},
	}

	r := httptest.NewRequest(http.MethodGet, "/v1/flows", nil)
	w := httptest.NewRecorder()
	GetFlows(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot find flows: something went wrong", logMessages[0].Message)
}

func TestGetFlow_ShouldReturnLatestVersionOfTheFlow(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		get: func(name string) (*Flow, error) {
			return &Flow{Name: name}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/flows/existingFlow?:flowName=existingFlow", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetFlow(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	expectedBody := `{"name":"existingFlow","links":[{"href":"http://example.com/v1/flows/existingFlow","rel":"self"},{"href":"http://example.com/v1/flows","rel":"up"},{"href":"http://example.com/swagger#/flow","rel":"help"}]}`
	assert.Equal(t, expectedBody, string(body))
}

func TestGetFlow_ShouldReturn404ForNonExistingFlow(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		get: func(name string) (*Flow, error) {
			return nil, FlowNotFoundErr
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/flow/nonExistingFlow", nil)
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
		get: func(name string) (*Flow, error) {
			return nil, errors.New("something went wrong")
		},
	}

	r := httptest.NewRequest(http.MethodGet, "/v1/flow/errorFlow", nil)
	w := httptest.NewRecorder()
	GetFlow(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot get flowName=%s: %vsomething went wrong", logMessages[0].Message)
}

func TestDeleteFlow_ShouldRemoveExistingFlow(t *testing.T) {

	defer resetFlowRepo()
	actualName := ""
	flowRepo = mockFlowRepo{
		remove: func(name string) error {
			actualName = name
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/flow/flowToDelete?:flowName=flowToDelete", nil)
	w := httptest.NewRecorder()
	DeleteFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "flowToDelete", actualName)
}

func TestDeleteFlow_ShouldReturn404ForNonExistingFlow(t *testing.T) {

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		remove: func(name string) error {
			return FlowNotFoundErr
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/flow/nonExistingFlow?:flowName=nonExistingFlow", nil)
	w := httptest.NewRecorder()
	DeleteFlow(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeleteFlow_ShouldReturn500_WhenErrorHappens(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetFlowRepo()
	flowRepo = mockFlowRepo{
		remove: func(name string) error {
			return errors.New("something went wrong")
		},
	}

	r := httptest.NewRequest(http.MethodDelete, "/v1/flow/errorFlow?:flowName=flowToDelete", nil)
	w := httptest.NewRecorder()
	DeleteFlow(w, r)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot delete flowName=flowToDelete: something went wrong", logMessages[0].Message)
}

// --- mocks & helpers ---

type mockFlowRepo struct {
	add     func(flow Flow) error
	remove  func(name string) error
	get     func(name string) (*Flow, error)
	findAll func() ([]Flow, error)
}

func resetFlowRepo() {
	flowRepo = flowMgoRepo{}
}

func (r mockFlowRepo) Add(flow Flow) error {
	return r.add(flow)
}

func (r mockFlowRepo) Remove(name string) error {
	return r.remove(name)
}

func (r mockFlowRepo) Get(name string) (*Flow, error) {
	return r.get(name)
}

func (r mockFlowRepo) FindAll() ([]Flow, error) {
	return r.findAll()
}

const redeployFlow = `{
    "name": "redeploy_flow",
    "description": "Redeploys app",
    "steps": [
        {
            "id" : "hipchat_start",
            "event": {
                "packName": "Hipchat",
                "packLabels": {
                    "env" : "staging"
                },
                "name": "MessageReceived"
            },
            "criteria": "{{ \"room1234\" == \"room1234\" }}",
            "context": {
                "roomId":"room1234"
            },
            "command": {
                "packName": "Argo",
                "name": "PutArtifact",
                "input": {
                    "id": "artifact5678",
                    "name": "foo-app",
                    "version": "1.0"
                }
            }
        },
        {
            "id" : "argo_to_hipchat",
            "dependsOn" : ["hipchat_start"],
            "event": {
                    "packName": "Argo",
                    "name": "ArtifactUpdated"
            },
            "command": {
                "packName": "Hipchat",
                "packLabels": {
                    "env" : "staging"
                },
                "name": "SendMessage",
                "input": {
                    "roomId" : "room1234",
                    "message": "Pipeline created"
                }
            }
        }
    ]
}`

const validJsonWithMissingField = `{
  "description": "Get some help on what you can do with argo and flyte",
  "steps": [
    {
      "id": "receive_help_message_send_response",
      "event": {
        "packName": "Slack",
        "name": "ReceivedMessage"
      },
      "context": {
        "Tts": "{% if Event.Payload.threadTimestamp != '' %}{{ Event.Payload.threadTimestamp }}{% else %}{{ Event.Payload.timestamp }}{% endif %}"
      },
      "criteria": "{{ Event.Payload.message|match:'^flyte(\\\\s+)help$' }}",
      "command": {
        "packName": "Slack",
        "name": "SendMessage",
        "input": {
          "channelId": "{{ Event.Payload.channelId }}",
          "threadTimestamp": "{{ Context.Tts }}",
          "message": "Hey <@{{vent.Payload.user.id }}>!"
        }
      }
    }
  ]
}`
