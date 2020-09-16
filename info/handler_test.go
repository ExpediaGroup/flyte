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

package info

import (
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexLinks(t *testing.T) {
	req := httptest.NewRequest("GET", "/", strings.NewReader(""))
	httputil.SetProtocolAndHostIn(req)

	responseWriter := httptest.NewRecorder()

	Index(responseWriter, req)

	assert.Equal(t, http.StatusOK, responseWriter.Code)
	assert.Equal(t, `{"links":[{"href":"http://example.com/","rel":"self"},`+
		`{"href":"http://example.com/swagger#/info","rel":"help"},`+
		`{"href":"http://example.com/v1","rel":"http://example.com/swagger#!/info/v1"}]}`,
		responseWriter.Body.String())
}

func TestV1Links(t *testing.T) {
	req := httptest.NewRequest("GET", "/v1", strings.NewReader(""))
	httputil.SetProtocolAndHostIn(req)
	responseWriter := httptest.NewRecorder()

	V1(responseWriter, req)

	assert.Equal(t, http.StatusOK, responseWriter.Code)
	assert.Equal(t, `{"links":[{"href":"http://example.com/v1","rel":"self"},`+
		`{"href":"http://example.com/","rel":"up"},`+
		`{"href":"http://example.com/swagger#!/info/v1","rel":"help"},`+
		`{"href":"http://example.com/health","rel":"http://example.com/swagger#!/info/health"},`+
		`{"href":"http://example.com/v1/packs","rel":"http://example.com/swagger#!/pack/listPacks"},`+
		`{"href":"http://example.com/v1/flows","rel":"http://example.com/swagger#!/flow/listFlows"},`+
		`{"href":"http://example.com/v1/datastore","rel":"http://example.com/swagger#!/datastore/listDatastoreItems"},`+
		`{"href":"http://example.com/v1/audit/flows","rel":"http://example.com/swagger#!/flowAudit/findFlows"},`+
		`{"href":"http://example.com/v1/swagger","rel":"http://example.com/swagger"}]}`, responseWriter.Body.String())
}

func TestSwagger_shouldReturnError_whenSwaggerFileCannotBeRead(t *testing.T) {
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	// conveniently, the 'ReadFile' code doesn't work when run in tests (see fix in test above) so an error will be returned...

	req := httptest.NewRequest("GET", "/v1/swagger", strings.NewReader(""))
	responseWriter := httptest.NewRecorder()

	V1Swagger(responseWriter, req)

	assert.Equal(t, http.StatusInternalServerError, responseWriter.Code)
	assert.Contains(t, loggertest.GetLogMessages()[0].Message, "cannot read swagger/v1.yml:")
}
