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

package flytepath

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUriFor_shouldReturnUriForTheNamePassedIn(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Proto = "http"
	request.Host = "www.example.com"

	EnsureUriDocMapIsInitialised(request)

	for key, value := range getFlyteDocPaths() {
		assert.Contains(t, GetUriDocPathFor(key), "http://www.example.com"+value)
	}
	resetPathMap()
}

func TestEnsureUriMapIsInitialisedWith_shouldInitialiseMapOnceAndOnlyOnce(t *testing.T) {
	// ensure uriMap is not initialised
	assert.Equal(t, 0, len(uriMap))

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Proto = "http"
	request.Host = "www.example.com"

	// 1st pass...
	EnsureUriDocMapIsInitialised(request)

	for key, value := range getFlyteDocPaths() {
		assert.Contains(t, GetUriDocPathFor(key), "http://www.example.com"+value)
	}

	// 2nd pass, different request proto and host
	request = httptest.NewRequest(http.MethodGet, "/", nil)
	request.Proto = "https"
	request.Host = "www.anotherexample.com"

	EnsureUriDocMapIsInitialised(request)

	// uriMap should not be changed from the initial initialisation
	for key, value := range getFlyteDocPaths() {
		assert.Contains(t, GetUriDocPathFor(key), "http://www.example.com"+value)
	}
	resetPathMap()
}

func resetPathMap() {
	uriMap = make(map[string]string)
}
