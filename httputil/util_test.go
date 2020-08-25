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

package httputil

import (
	"encoding/json"
	"errors"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWriteResponse_shouldWriteJsonResponse(t *testing.T) {

	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	WriteResponse(w, r, mockedInterface)

	body, err := ioutil.ReadAll(w.Result().Body)
	require.NoError(t, err)
	var got interface{}
	err = json.Unmarshal(body, &got)
	require.NoError(t, err)

	assert.Equal(t, mockedInterface, got)
	assert.Equal(t, ContentTypeJson, w.HeaderMap.Get(HeaderContentType))
}

func TestWriteResponse_shouldWriteYamlResponse_whenRequestHeaderAcceptIsYaml(t *testing.T) {
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderAccept, MediaTypeYaml)
	WriteResponse(w, r, mockedInterface)

	body, err := ioutil.ReadAll(w.Result().Body)
	require.NoError(t, err)
	var got interface{}
	err = yaml.Unmarshal(body, &got)
	require.NoError(t, err)

	assert.Equal(t, mockedInterface, got)
	assert.Equal(t, ContentTypeYaml, w.HeaderMap.Get(HeaderContentType))
}

func TestWriteResponse_shouldProduce500Response_whenUnableToMarshalJson(t *testing.T) {
	loggertest.Init(loggertest.LogLevelError)
	defer loggertest.Reset()
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	unmarshalableJson := math.NaN()
	WriteResponse(w, r, unmarshalableJson)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Contains(t, logMessages[0].Message, "cannot convert to JSON:")
}

func TestWriteResponse_shouldProduce500Response_whenUnableToMarshalYaml(t *testing.T) {
	loggertest.Init(loggertest.LogLevelError)
	defer loggertest.Reset()
	w := httptest.NewRecorder()

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set(HeaderAccept, MediaTypeYaml)
	unmarshalableYaml := math.NaN()
	WriteResponse(w, r, unmarshalableYaml)

	resp := w.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Contains(t, logMessages[0].Message, "cannot convert to yaml:")
}

func TestWriteResponse_shouldProduce500Response_whenUnableToWriteData(t *testing.T) {
	loggertest.Init(loggertest.LogLevelError)
	defer loggertest.Reset()

	expectedError := errors.New("some error")
	w := &mockResponseRecorder{expectedError: expectedError}
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	WriteResponse(w, r, "doesn't matter")

	assert.Equal(t, http.StatusInternalServerError, w.code)
	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, expectedError.Error(), logMessages[0].Message)
}

// -- mock functions & variables

var mockedInterface = map[string]interface{}{
	"name":      "John Doe",
	"job":       nil,
	"employed":  true,
	"foods":     []interface{}{"Apple", "Orange", "Strawberry", "Mango"},
	"languages": map[string]interface{}{"perl": "Elite", "go": "Elite", "pascal": "Lame"},
}

type mockResponseRecorder struct {
	code          int
	expectedError error
}

func (m *mockResponseRecorder) Write([]byte) (int, error) {
	return 0, m.expectedError
}

func (m *mockResponseRecorder) WriteHeader(code int) {
	m.code = code
}

func (m *mockResponseRecorder) Header() http.Header {
	return http.Header{}
}
