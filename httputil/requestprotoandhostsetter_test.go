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
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetUri_shouldReturnANonSecureUriBuiltFromTheRequestProtocolAndHost(t *testing.T) {
	request := httptest.NewRequest("GET", "http://localhost", strings.NewReader(""))

	SetProtocolAndHostIn(request)

	assert.Equal(t, "http", request.Proto)
	assert.Equal(t, "localhost", request.Host)
}

func TestGetUri_shouldReturnASecureUriBuiltFromTheRequestProtocolAndHost(t *testing.T) {
	request := httptest.NewRequest("GET", "https://localhost", strings.NewReader(""))

	SetProtocolAndHostIn(request)

	assert.Equal(t, "https", request.Proto)
	assert.Equal(t, "localhost", request.Host)
}

func TestGetUri_shouldReturnAUriBuiltFromAnXForwardedProtocolHeader(t *testing.T) {
	request := httptest.NewRequest("GET", "http://localhost", strings.NewReader(""))
	request.Header.Set("X-Forwarded-Proto", "https")

	SetProtocolAndHostIn(request)

	assert.Equal(t, "https", request.Proto)
	assert.Equal(t, "localhost", request.Host)
}

func TestGetUri_shouldReturnAUriBuiltFromAnXForwardedHostHeader(t *testing.T) {
	request := httptest.NewRequest("GET", "http://localhost", strings.NewReader(""))
	request.Header.Set("X-Forwarded-Host", "example.com")

	SetProtocolAndHostIn(request)

	assert.Equal(t, "http", request.Proto)
	assert.Equal(t, "example.com", request.Host)
}

func TestGetUri_shouldReturnAUriBuiltFromAnXForwardedProtocolAndHostHeader(t *testing.T) {
	request := httptest.NewRequest("GET", "http://localhost", strings.NewReader(""))
	request.Header.Set("X-Forwarded-Proto", "https")
	request.Header.Set("X-Forwarded-Host", "example.com")

	SetProtocolAndHostIn(request)

	assert.Equal(t, "https", request.Proto)
	assert.Equal(t, "example.com", request.Host)
}
