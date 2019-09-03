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
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var handlerToWrap = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "Hello, is it me you're looking for?")
})

func TestNewRequestInterceptor_ShouldWorkOutAndSetTheCorrectProtocolAndHostInTheRequest_BeforeFallingThroughToTheHandlerItHasWrapped(t *testing.T) {
	interceptor := NewRequestInterceptor(handlerToWrap)
	r := httptest.NewRequest(http.MethodGet, "http://somewhere.com/anypath", nil)
	r.Header.Set("X-Forwarded-Proto", "https")
	r.Header.Set("X-Forwarded-Host", "original-host")
	w := httptest.NewRecorder()

	assert.Equal(t, "HTTP/1.1", r.Proto)
	assert.Equal(t, "somewhere.com", r.Host)

	interceptor.ServeHTTP(w, r)

	assert.Equal(t, "https", r.Proto)
	assert.Equal(t, "original-host", r.Host)

	// assertions for wrapped interceptor
	resp := w.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "Hello, is it me you're looking for?\n", string(b))
}
