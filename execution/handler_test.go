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
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestPostEvent_ShouldHandleValidEvent(t *testing.T) {

	//Given
	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			if id == "Slack" {
				return &Pack{Id: "Slack"}, nil
			}
			return nil, PackNotFoundErr
		},
	}

	defer resetFlowService()
	var wg sync.WaitGroup
	wg.Add(1)
	actualEvent := Event{}
	flowSvc = mockFlowService{
		handleEvent: func(e Event) {
			actualEvent = e
			wg.Done()
		},
	}

	//When
	w := httptest.NewRecorder()
	PostEvent(w, httptest.NewRequest(http.MethodPost, "/v1/packs/Slack/events?:packId=Slack", eventBody()))

	//Then
	resp := w.Result()
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	waitWithTimeout(wg, 500*time.Millisecond)
	expectedEvent := Event{Name: "MessageReceived", Pack: Pack{Id: "Slack"}, Payload: map[string]interface{}{"channelId": "123456"}}
	assert.Equal(t, expectedEvent, actualEvent)
}

func TestPostEvent_ShouldReturn404WhenPackDoesNotExist(t *testing.T) {

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return nil, PackNotFoundErr
		},
	}

	w := httptest.NewRecorder()
	PostEvent(w, httptest.NewRequest(http.MethodPost, "/v1/packs/Slack/events?:packId=Slack", eventBody()))

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestPostEvent_ShouldReturn500WhenThereIsErrorRetrievingPack(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return nil, errors.New("it is an error")
		},
	}

	w := httptest.NewRecorder()
	PostEvent(w, httptest.NewRequest(http.MethodPost, "/v1/packs/Slack/events?:packId=Slack", eventBody()))

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "it is an error", logMessages[0].Message)
}

func TestPostEvent_ShouldReturn400WhenRequestBodyIsInvalid(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return &Pack{Id: "Slack"}, nil
		},
	}

	w := httptest.NewRecorder()
	PostEvent(w, httptest.NewRequest(http.MethodPost,
		"/v1/packs/Slack/events?:packId=Slack", strings.NewReader(`{"invalidBody`)))

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "unexpected EOF", logMessages[0].Message)
}

func TestCompleteAction_ShouldCompleteActionAndHandleIt(t *testing.T) {

	//Given
	defer resetPackRepo()
	pack := Pack{Id: "Slack"}
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return &Pack{Id: id}, nil
		},
	}

	defer resetCompleteAction()
	completeAction = func(p Pack, actionId string, result Event) (*Action, error) {
		if p.Id == pack.Id && actionId == "123" {
			return &Action{Id: actionId, PackName: p.Name, Result: result, State: State{Value: stateSuccess}}, nil
		}
		t.Fatal("Should not get here")
		return nil, nil
	}

	defer resetFlowService()
	var wg sync.WaitGroup
	wg.Add(2)
	actualEvent := Event{}
	actualAction := Action{}
	flowSvc = mockFlowService{
		handleEvent: func(e Event) {
			actualEvent = e
			wg.Done()
		},
		handleAction: func(a Action) {
			actualAction = a
			wg.Done()
		},
	}

	//When
	w := httptest.NewRecorder()
	CompleteAction(w, httptest.NewRequest(http.MethodPost,
		"/v1/packs/Slack/actions/123/result?:packId=Slack&:actionId=123", eventBody()))

	//Then
	resp := w.Result()
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	waitWithTimeout(wg, 500*time.Millisecond)
	event := Event{Name: "MessageReceived", Pack: pack, Payload: map[string]interface{}{"channelId": "123456"}}
	action := Action{Id: "123", PackName: pack.Name, Result: event, State: State{Value: stateSuccess}}
	assert.Equal(t, event, actualEvent)
	assert.Equal(t, action, actualAction)
}

func TestCompleteAction_ShouldReturn404WhenPackDoesNotExist(t *testing.T) {

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return nil, PackNotFoundErr
		},
	}

	w := httptest.NewRecorder()
	CompleteAction(w, httptest.NewRequest(http.MethodPost,
		"/v1/packs/Slack/actions/123/result?:packId=Slack&:actionId=123", eventBody()))

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCompleteAction_ShouldReturn500WhenThereIsErrorRetrievingPack(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return nil, errors.New("it is an error")
		},
	}

	w := httptest.NewRecorder()
	CompleteAction(w, httptest.NewRequest(http.MethodPost,
		"/v1/packs/Slack/actions/123/result?:packId=Slack&:actionId=123", eventBody()))

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "it is an error", logMessages[0].Message)
}

func TestCompleteAction_ShouldReturn400WhenRequestBodyIsInvalid(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return &Pack{Id: id}, nil
		},
	}

	w := httptest.NewRecorder()
	CompleteAction(w, httptest.NewRequest(http.MethodPost,
		"/v1/packs/Slack/actions/123/result?:packId=Slack&:actionId=123", strings.NewReader(`{"invalidBody`)))

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "unexpected EOF", logMessages[0].Message)
}

func TestCompleteAction_ShouldReturn404ForNonExistingAction(t *testing.T) {

	//Given
	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return &Pack{Id: id}, nil
		},
	}

	defer resetCompleteAction()
	completeAction = func(pack Pack, actionId string, result Event) (*Action, error) {
		if pack.Id == "Slack" && actionId == "123" {
			return nil, ActionNotFoundErr
		}
		t.Fatal("Should not get here")
		return nil, nil
	}

	//When
	w := httptest.NewRecorder()
	CompleteAction(w, httptest.NewRequest(http.MethodPost,
		"/v1/packs/Slack/actions/123/result?:packId=Slack&:actionId=123", eventBody()))

	//Then
	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestCompleteAction_ShouldReturn500WhenThereIsErrorCompletingAction(t *testing.T) {

	//Given
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return &Pack{Id: id}, nil
		},
	}

	defer resetCompleteAction()
	completeAction = func(pack Pack, actionId string, result Event) (*Action, error) {
		if pack.Id == "Slack" && actionId == "123" {
			return nil, errors.New("it's a disaster, run run run")
		}
		t.Fatal("Should not get here")
		return nil, nil
	}

	//When
	w := httptest.NewRecorder()
	CompleteAction(w, httptest.NewRequest(http.MethodPost,
		"/v1/packs/Slack/actions/123/result?:packId=Slack&:actionId=123", eventBody()))

	//Then
	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Error completing actionId=123 with result=&{Name:MessageReceived Pack:{Id:Slack Name: Labels:map[]} Payload:map[channelId:123456]}: it's a disaster, run run run", logMessages[0].Message)
}

func TestTakeAction_ShouldReturnActionWhenPackHasAnyNewActionsAndNameIsNotSpecified(t *testing.T) {

	//Given
	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return &Pack{Id: id}, nil
		},
	}

	defer resetTakeAction()
	takeAction = func(p Pack, actionName string) (*Action, error) {
		if p.Id == "Slack" && actionName == "" {
			return &Action{Id: "596759ef", PackName: p.Name, Name: actionName}, nil
		}
		t.Fatal("Should not get here")
		return nil, nil
	}

	//When
	w := httptest.NewRecorder()

	request := httptest.NewRequest(http.MethodPost, "/v1/packs/Slack/actions/take?:packId=Slack", nil)
	httputil.SetProtocolAndHostIn(request)
	flytepath.EnsureUriDocMapIsInitialised(request)
	TakeAction(w, request)

	//Then
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	expectedBody := `{"command":"","input":null,"links":[{"href":"http://example.com/v1/packs/Slack/actions/596759ef/result","rel":"http://example.com/swagger#/actionResult"}]}`
	assert.Equal(t, expectedBody, string(body))
}

func TestTakeAction_ShouldReturnActionWhenPackHasNewActionsWithTheGivenName(t *testing.T) {

	//Given
	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return &Pack{Id: id}, nil
		},
	}

	defer resetTakeAction()
	takeAction = func(p Pack, actionName string) (*Action, error) {
		if p.Id == "Slack" && actionName == "SendMessage" {
			return &Action{Id: "596759ef", PackName: p.Name, Name: actionName}, nil
		}
		t.Fatal("Should not get here")
		return nil, nil
	}

	//When
	w := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/v1/packs/Slack/actions/take?:packId=Slack&actionName=SendMessage", nil)
	httputil.SetProtocolAndHostIn(request)
	flytepath.EnsureUriDocMapIsInitialised(request)
	TakeAction(w, request)

	//Then
	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	expectedBody := `{"command":"SendMessage","input":null,"links":[{"href":"http://example.com/v1/packs/Slack/actions/596759ef/result","rel":"http://example.com/swagger#/actionResult"}]}`
	assert.Equal(t, expectedBody, string(body))
}

func TestTakeAction_ShouldReturn404WhenPackDoesNotExist(t *testing.T) {

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return nil, PackNotFoundErr
		},
	}

	w := httptest.NewRecorder()
	TakeAction(w, httptest.NewRequest(http.MethodPost, "/v1/packs/Slack/actions/take?:packId=Slack", nil))

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestTakeAction_ShouldReturn500WhenThereIsErrorRetrievingPack(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return nil, errors.New("it is an error")
		},
	}

	w := httptest.NewRecorder()
	TakeAction(w, httptest.NewRequest(http.MethodPost, "/v1/packs/Slack/actions/take?:packId=Slack", nil))

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "it is an error", logMessages[0].Message)
}

func TestTakeAction_ShouldReturn500WhenThereIsErrorTakingAction(t *testing.T) {

	//Given
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return &Pack{Id: id}, nil
		},
	}

	defer resetTakeAction()
	takeAction = func(p Pack, actionName string) (*Action, error) {
		if p.Id == "Slack" && actionName == "" {
			return nil, errors.New("it's a disaster, run run run")
		}
		t.Fatal("Should not get here")
		return nil, nil
	}

	//When
	w := httptest.NewRecorder()
	TakeAction(w, httptest.NewRequest(http.MethodPost, "/v1/packs/Slack/actions/take?:packId=Slack", nil))

	//Then
	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Could not take action for packId=Slack and actionName=: it's a disaster, run run run", logMessages[0].Message)
}

func TestTakeAction_ShouldReturn204WhenPackDoesNotHaveNewAction(t *testing.T) {

	//Given
	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return &Pack{Id: id}, nil
		},
	}

	defer resetTakeAction()
	takeAction = func(p Pack, actionName string) (*Action, error) {
		if p.Id == "Slack" && actionName == "" {
			return nil, nil
		}
		t.Fatal("Should not get here")
		return nil, nil
	}

	//When
	w := httptest.NewRecorder()
	TakeAction(w, httptest.NewRequest(http.MethodPost, "/v1/packs/Slack/actions/take?:packId=Slack", nil))

	//Then
	resp := w.Result()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// --- mocks & helpers ---

type mockFlowService struct {
	handleEvent  func(e Event)
	handleAction func(a Action)
}

func (s mockFlowService) HandleEvent(e Event) {
	s.handleEvent(e)
}

func (s mockFlowService) HandleAction(a Action) {
	s.handleAction(a)
}

type mockPackRepo struct {
	get func(id string) (*Pack, error)
}

func (r mockPackRepo) Get(id string) (*Pack, error) {
	return r.get(id)
}

func resetFlowService()    { flowSvc = flowService{} }
func resetPackRepo()       { packRepo = packMgoRepo{} }
func resetCompleteAction() { completeAction = completeActionFn }
func resetTakeAction()     { takeAction = takeActionFn }

func eventBody() io.Reader {
	return strings.NewReader(`{"event": "MessageReceived", "payload": {"channelId": "123456"}}`)
}
