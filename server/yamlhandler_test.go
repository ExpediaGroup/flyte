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
	"encoding/json"
	"github.com/husobee/vestigo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"strings"
	"testing"
	"errors"
)

func TestYamlHandler_shouldParseYamlIntoJSON(t *testing.T) {

	var gotJSON interface{}
	var modifiedReq *http.Request
	// the yaml handler will decorate this handler
	downstreamHandler := func(rw http.ResponseWriter, r *http.Request) {
		modifiedReq = r
		err := json.NewDecoder(r.Body).Decode(&gotJSON)
		require.NoError(t, err)
	}

	router := vestigo.NewRouter()
	router.Post("/", downstreamHandler, YamlHandler)
	server := httptest.NewServer(router)
	defer server.Close()

	_, err := http.DefaultClient.Post(server.URL+"/", httputil.MediaTypeYaml, strings.NewReader(multipleDataTypesYaml))
	require.NoError(t, err)

	var expectedJSON interface{}
	err = json.Unmarshal([]byte(multipleDataTypesJSON), &expectedJSON)
	require.NoError(t, err)
	assert.Equal(t, expectedJSON, gotJSON)

	b := new(bytes.Buffer)
	err = json.Compact(b, []byte(multipleDataTypesJSON))
	require.NoError(t, err)
	assert.Equal(t, int64(b.Len()), modifiedReq.ContentLength)
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

const multipleDataTypesYaml = `
---
# An employee record
name: "John Doe"
job: Developer
skill: &myVar Elite
employed: True
foods:
    - Apple
    - Orange
    - Strawberry
    - Mango
languages:
    perl: *myVar
    go: *myVar
    pascal: Lame
education: |
    4 GCSEs
    3 A-Levels
    BSc in the Internet of Things
`

const multipleDataTypesJSON = `{
  "name": "John Doe",
  "job": "Developer",
  "skill": "Elite",
  "employed": true,
  "foods": [
    "Apple",
    "Orange",
    "Strawberry",
    "Mango"
  ],
  "languages": {
    "perl": "Elite",
    "go": "Elite",
    "pascal": "Lame"
  },
  "education": "4 GCSEs\n3 A-Levels\nBSc in the Internet of Things\n"
}`

type mockIoReader struct {}

func (r *mockIoReader) Read([]byte) (int, error) {
	return 0, errors.New("mockIoReader error")
}
