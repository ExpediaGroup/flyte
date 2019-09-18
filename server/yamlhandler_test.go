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

package server

import (
	"bytes"
	"errors"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/husobee/vestigo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestYamlHandler_shouldParseYamlIntoJSON(t *testing.T) {
	got := new(bytes.Buffer)
	// the yaml handler will decorate this handler
	downstreamHandler := func(rw http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		_, err := got.ReadFrom(r.Body)
		require.NoError(t, err)
	}

	router := vestigo.NewRouter()
	router.Post("/", downstreamHandler, YamlHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	_, err := http.DefaultClient.Post(server.URL+"/", httputil.MediaTypeYaml, strings.NewReader(validYaml))
	require.NoError(t, err)

	assert.JSONEq(t, multipleDataTypesJSON, got.String())
}

func TestYamlHandler_shouldReturnBadRequest_whenRequestContainsInvalidYaml(t *testing.T) {
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)
	numInvocations := 0
	// the yaml handler will decorate this handler
	downstreamHandler := func(rw http.ResponseWriter, r *http.Request) { numInvocations++ }

	router := vestigo.NewRouter()
	router.Post("/", downstreamHandler, YamlHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	_, err := http.DefaultClient.Post(server.URL+"/", httputil.MediaTypeYaml, strings.NewReader("bad: yaml \nfoo"))
	require.NoError(t, err)

	assert.Equal(t, numInvocations, 0)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Contains(t, logMessages[0].Message, "cannot process yaml request:")
}

func TestConvertYAMLRequestToJSONRequest_shouldError_whenRequestCannotBeRead(t *testing.T) {
	mockReader := &mockIoReader{}
	req, err := http.NewRequest(http.MethodPost, "/", mockReader)
	require.NoError(t, err)

	err = convertYAMLRequestToJSONRequest(req)
	assert.Error(t, err)
	assert.Equal(t, "mockIoReader error", err.Error())
}

const multipleDataTypesJSON = `{
  "name": "validYaml",
  "description": "Simple flow to test if flyte & slack pack are up and running.",
  "steps": [
    {
      "event": {
        "packName": "Slack",
        "name": "ReceivedMessage"
      },
      "context": {
        "ChannelID": "{{ Event.Payload.channelId }}",
        "UserID": "{{ Event.Payload.user.id }}",
        "Tts": "{% if Event.Payload.threadTimestamp != '' %}{{ Event.Payload.threadTimestamp }}{% else %}{{ Event.Payload.timestamp }}{% endif %}"
      },
      "command": {
        "name": "SendMessage",
        "packName": "Slack"
			}
    }
  ]
}
`

var invalidYaml = `
description: Simple flow to test if flyte & slack pack are up and running.
`

var validYaml = `
description: "Simple flow to test if flyte & slack pack are up and running."
name: validYaml
steps:
  - 
    context: 
      ChannelID: "{{ Event.Payload.channelId }}"
      Tts: "{% if Event.Payload.threadTimestamp != '' %}{{ Event.Payload.threadTimestamp }}{% else %}{{ Event.Payload.timestamp }}{% endif %}"
      UserID: "{{ Event.Payload.user.id }}"
    event: 
      name: ReceivedMessage
      packName: Slack
    command:
      packName: Slack
      name: SendMessage
`

type mockIoReader struct{}

func (r *mockIoReader) Read([]byte) (int, error) {
	return 0, errors.New("mockIoReader error")
}
