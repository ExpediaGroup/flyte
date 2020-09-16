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

package pack

import (
	"encoding/json"
	"errors"
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPostPack_ShouldCreatePackForValidRequest(t *testing.T) {

	defer resetPackRepo()
	packRepo = mockPackRepo{
		add: func(pack Pack) error {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/packs", strings.NewReader(slackPackJson))
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	PostPack(w, req)

	resp := w.Result()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	location, err := resp.Location()
	require.NoError(t, err)
	assert.Equal(t, "http://example.com/v1/packs/Slack", location.String())

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	var got Pack
	err = json.Unmarshal(body, &got)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), got.LastSeen, 5*time.Second)

	lsB, err := json.Marshal(got.LastSeen)
	require.NoError(t, err)
	packResp := strings.Replace(slackPackResponse, "replace_last_seen", string(lsB), 1)

	assert.JSONEq(t, packResp, string(body))
}

func TestPostPack_should_fail_with_bad_request(t *testing.T) {
	defer resetPackRepo()
	packRepo = mockPackRepo{
		add: func(pack Pack) error {
			return nil
		},
	}

	tests := []struct {
		name string
		link httputil.Link
	}{
		{
			name: "'self' is used as a relative name",
			link: httputil.Link{Href: "http://somewhere.com", Rel: "self"},
		},
		{
			name: "'up' is used as a relative name",
			link: httputil.Link{Href: "http://somewhere.com", Rel: "up"},
		},
		{
			name: "'Rel' attribute ends with 'actionResult'",
			link: httputil.Link{Href: "http://somewhere.com", Rel: "somewhere.com/actionResult"},
		},
		{
			name: "'Rel' attribute ends with 'takeAction'",
			link: httputil.Link{Href: "http://somewhere.com", Rel: "somewhere.com/takeAction"},
		},
		{
			name: "'Rel' attribute ends with '/event'",
			link: httputil.Link{Href: "http://somewhere.com", Rel: "somewhere.com/something/event"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pack := Pack{
				Name: "Slack",
				Commands: []Command{
					{Name: "SendMessage", Events: []string{"MessageSent", "SendMessageFailed"}},
				},
				Events: []Event{
					{Name: "MessageSent"},
					{Name: "SendMessageFailed"},
				},
				Links: []httputil.Link{
					{Href: "http://example.com/README.md"},
				},
			}

			pack.Links = append(pack.Links, test.link)

			bytes, err := json.Marshal(&pack)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/v1/packs", strings.NewReader(string(bytes)))
			httputil.SetProtocolAndHostIn(req)
			w := httptest.NewRecorder()
			PostPack(w, req)

			resp := w.Result()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestPostPack_ShouldReturn400ForInvalidRequest(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	req := httptest.NewRequest(http.MethodPost, "/v1/packs", strings.NewReader(`--- invalid json ---`))
	w := httptest.NewRecorder()
	PostPack(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot convert request to pack: invalid character '-' in numeric literal", logMessages[0].Message)
}

func TestPostPack_ShouldReturn500_WhenRepoFails(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		add: func(pack Pack) error {
			return errors.New("something went wrong")
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/v1/packs", strings.NewReader(slackPackJson))
	w := httptest.NewRecorder()
	PostPack(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot save packName=Slack, packLabels=map[]: something went wrong", logMessages[0].Message)
}

func TestGetPacks_ShouldReturnListOfPacksWithLinks_WhenPacksExist(t *testing.T) {

	defer resetPackRepo()
	tLive := time.Now()
	tWarning := tLive.Add(-1 * time.Hour)

	packRepo = mockPackRepo{
		findAll: func() ([]Pack, error) {
			slack := Pack{Id: "Slack", Name: "Slack", Labels: map[string]string{"env": "dev"}, LastSeen: tLive}
			hipChat := Pack{Id: "HipChat", Name: "HipChat", LastSeen: tWarning}
			return []Pack{slack, hipChat}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/packs", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetPacks(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))

	tLiveJson, err := tLive.MarshalJSON()
	require.NoError(t, err)
	packsResp := strings.Replace(slackAndHipchatPacksResponse, "replace_last_seen_slack", string(tLiveJson), 1)
	tWarningJson, err := tWarning.MarshalJSON()
	require.NoError(t, err)
	packsResp = strings.Replace(packsResp, "replace_last_seen_hipchat", string(tWarningJson), 1)

	assert.JSONEq(t, packsResp, string(body))
}

func TestGetPacks_ShouldReturnEmptyListOfPacksWithLinks_WhenThereAreNoPacks(t *testing.T) {

	defer resetPackRepo()
	packRepo = mockPackRepo{
		findAll: func() ([]Pack, error) {
			return []Pack{}, nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/packs", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetPacks(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, resp.Header.Get(httputil.HeaderContentType))
	assert.Equal(t, emptyPacksResponse, string(body))
}

func TestGetPacks_ShouldReturn500_WhenRepoFails(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		findAll: func() ([]Pack, error) {
			return nil, errors.New("something went wrong")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/packs", nil)
	w := httptest.NewRecorder()
	GetPacks(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot find packs: something went wrong", logMessages[0].Message)
}

func TestGetPack_ShouldReturnPack(t *testing.T) {

	defer resetPackRepo()
	now := time.Now()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			if id == "Slack" {
				slack := &Pack{}
				json.NewDecoder(strings.NewReader(slackPackJson)).Decode(slack)
				slack.Id = "Slack"
				slack.LastSeen = now
				return slack, nil
			}
			return nil, PackNotFoundErr
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/packs/Slack?:packId=Slack", nil)
	httputil.SetProtocolAndHostIn(req)
	w := httptest.NewRecorder()
	GetPack(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.Nil(t, err)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	nowJson, err := now.MarshalJSON()
	require.NoError(t, err)
	packResp := strings.Replace(slackPackResponse, "replace_last_seen", string(nowJson), 1)
	assert.JSONEq(t, packResp, string(body))
}

func TestGetPack_Should404ForNonExistingPack(t *testing.T) {

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return nil, PackNotFoundErr
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/packs/Slack", nil)
	w := httptest.NewRecorder()
	GetPack(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetPack_Should500_WhenRepoFails(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		get: func(id string) (*Pack, error) {
			return nil, errors.New("something went wrong")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/packs/Slack", nil)
	w := httptest.NewRecorder()
	GetPack(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot find packId=: something went wrong", logMessages[0].Message)
}

func TestDeletePack_ShouldDeleteExistingPack(t *testing.T) {

	defer resetPackRepo()
	packRepo = mockPackRepo{
		remove: func(id string) error {
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/packs/Slack", nil)
	w := httptest.NewRecorder()
	DeletePack(w, req)

	resp := w.Result()
	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Empty(t, string(body))
}

func TestDeletePack_Should404ForNonExistingPack(t *testing.T) {

	defer resetPackRepo()
	packRepo = mockPackRepo{
		remove: func(id string) error {
			return PackNotFoundErr
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/packs/Slack", nil)
	w := httptest.NewRecorder()
	DeletePack(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestDeletePack_Should500_WhenRepoFails(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetPackRepo()
	packRepo = mockPackRepo{
		remove: func(id string) error {
			return errors.New("something went wrong")
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/v1/packs/Slack", nil)
	w := httptest.NewRecorder()
	DeletePack(w, req)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot delete packId=: something went wrong", logMessages[0].Message)
}

// --- requests/responses ---

var slackPackJson = `
{
    "name": "Slack",
    "commands": [
        {
            "name": "SendMessage",
            "events": ["MessageSent", "SendMessageFailed"]
        }
    ],
    "events": [
        {
            "name": "MessageSent"
        },
        {
            "name": "SendMessageFailed"
        }
    ],
	"links": [
		{
			"href": "http://example.com/README.md",
			"rel": "help"
		}
	]
}
`

var slackPackResponse = strings.Replace(strings.Replace(`
{
    "id": "Slack",
    "name": "Slack",
    "commands": [
        {
            "name": "SendMessage",
            "events": ["MessageSent", "SendMessageFailed"],
            "links": [
                {
                    "href": "http://example.com/v1/packs/Slack/actions/take?commandName=SendMessage",
                    "rel": "http://example.com/swagger#!/action/takeAction"
                }
            ]
        }
    ],
    "events": [
        {
            "name": "MessageSent"
        },
        {
            "name": "SendMessageFailed"
        }
    ],
    "lastSeen": replace_last_seen,
    "status": "live",
    "links": [
        {
            "href": "http://example.com/README.md",
            "rel": "help"
        },
        {
            "href": "http://example.com/v1/packs/Slack",
            "rel": "self"
        },
        {
            "href": "http://example.com/v1/packs",
            "rel": "up"
        },
        {
            "href": "http://example.com/v1/packs/Slack/actions/take",
            "rel": "http://example.com/swagger#!/action/takeAction"
        },
        {
            "href": "http://example.com/v1/packs/Slack/events",
            "rel": "http://example.com/swagger#/event"
        }
    ]
}
`, "\n", "", -1), " ", "", -1)

var slackAndHipchatPacksResponse = strings.Replace(strings.Replace(`
{
    "links": [
        {
            "href": "http://example.com/v1/packs",
            "rel": "self"
        },
        {
            "href": "http://example.com/v1",
            "rel": "up"
        },
        {
            "href": "http://example.com/swagger#/pack",
            "rel": "help"
        }
    ],
    "packs": [
        {
            "id": "Slack",
            "name": "Slack",
            "labels": {
                "env": "dev"
            },
            "lastSeen": replace_last_seen_slack,
			"status": "live",
            "links": [
                {
                    "href": "http://example.com/v1/packs/Slack",
                    "rel": "self"
                }
            ]
        },
        {
            "id": "HipChat",
            "name": "HipChat",
            "lastSeen": replace_last_seen_hipchat,
			"status": "warning",
            "links": [
                {
                    "href": "http://example.com/v1/packs/HipChat",
                    "rel": "self"
                }
            ]
        }
    ]
}
`, "\n", "", -1), " ", "", -1)

var emptyPacksResponse = strings.Replace(strings.Replace(`
{
    "links": [
        {
            "href": "http://example.com/v1/packs",
            "rel": "self"
        },
        {
            "href": "http://example.com/v1",
            "rel": "up"
        },
        {
            "href": "http://example.com/swagger#/pack",
            "rel": "help"
        }
    ],
    "packs": []
}
`, "\n", "", -1), " ", "", -1)

// --- mocks & helpers ---

type mockPackRepo struct {
	add                func(pack Pack) error
	remove             func(id string) error
	get                func(id string) (*Pack, error)
	findAll            func() ([]Pack, error)
	removeAllOlderThan func(date time.Time) (packsRemoved int, err error)
}

func (r mockPackRepo) Add(pack Pack) error {
	return r.add(pack)
}

func (r mockPackRepo) Remove(id string) error {
	return r.remove(id)
}

func (r mockPackRepo) Get(id string) (*Pack, error) {
	return r.get(id)
}

func (r mockPackRepo) FindAll() ([]Pack, error) {
	return r.findAll()
}

func (r mockPackRepo) RemoveAllOlderThan(date time.Time) (packsRemoved int, err error) {
	return r.removeAllOlderThan(date)
}

func resetPackRepo() {
	packRepo = packMgoRepo{}
}
