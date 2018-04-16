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
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUriBuilder_shouldReturnUriBuiltFromFromRequestProtocolAndHost(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Proto = "http"
	request.Host = "www.example.com"

	builder := UriBuilder(request)
	uri := builder.Build()

	assert.Equal(t, "http://www.example.com/", uri)
}

func TestUriBuilder_shouldReturnUriWithPath(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Proto = "http"
	request.Host = "www.example.com"

	builder := UriBuilder(request)
	builder.Path("/packs", "/hipchat")
	uri := builder.Build()

	assert.Equal(t, "http://www.example.com/packs/hipchat", uri)
}

func TestUriBuilder_shouldReturnUriWithParentPath(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Proto = "http"
	request.Host = "www.example.com"

	builder := UriBuilder(request)
	builder.Path("/packs")
	builder.Parent()
	uri := builder.Build()

	assert.Equal(t, "http://www.example.com/", uri)
}

func TestUriBuilder_shouldReturnUriWithParentPathWithNoTrailingSlash(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Proto = "http"
	request.Host = "www.example.com"

	builder := UriBuilder(request)
	builder.Path("/")
	builder.Parent()
	uri := builder.Build()

	assert.Equal(t, "http://www.example.com/", uri)
}

func TestUriBuilder_shouldReturnUriWithPathParameterReplacedWithCorrectValue(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Proto = "http"
	request.Host = "www.example.com"

	builder := UriBuilder(request)
	builder.Path("/packs/:pack")
	builder.Replace(":pack", "hipchat")
	uri := builder.Build()

	assert.Equal(t, "http://www.example.com/packs/hipchat", uri)
}

func TestUriBuilder_shouldReturnUriWithPathParameterRemovedIfValueIsEmpty(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Proto = "http"
	request.Host = "www.example.com"

	builder := UriBuilder(request)
	builder.Path("/packs/:pack")
	builder.Replace(":pack", "")
	uri := builder.Build()

	assert.Equal(t, "http://www.example.com/packs", uri)
}
