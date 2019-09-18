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
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/docker/docker/pkg/testutil/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostingPack_shouldProcessRequestThroughHandler(t *testing.T) {
	numInvocations := 0
	cleanupFunc := mockYamlHandler(
		func(http.ResponseWriter, *http.Request) { numInvocations++ },
	)
	defer cleanupFunc()

	server := httptest.NewServer(Handler())
	defer server.Close()

	_, err := http.DefaultClient.Post(server.URL+flytepath.PacksPath, "any content type", nil)
	require.NoError(t, err)

	assert.Equal(t, numInvocations, 1)
}

func TestPostingTakeAction_shouldProcessRequestThroughHandler(t *testing.T) {
	numInvocations := 0
	cleanupFunc := mockYamlHandler(
		func(http.ResponseWriter, *http.Request) { numInvocations++ },
	)
	defer cleanupFunc()

	server := httptest.NewServer(Handler())
	defer server.Close()

	_, err := http.DefaultClient.Post(server.URL+flytepath.TakeActionPath, "any content type", nil)
	require.NoError(t, err)

	assert.Equal(t, numInvocations, 1)
}

func TestPostingEvent_shouldProcessRequestThroughHandler(t *testing.T) {
	numInvocations := 0
	cleanupFunc := mockYamlHandler(
		func(http.ResponseWriter, *http.Request) { numInvocations++ },
	)
	defer cleanupFunc()

	server := httptest.NewServer(Handler())
	defer server.Close()

	_, err := http.DefaultClient.Post(server.URL+flytepath.PostEventPath, "any content type", nil)
	require.NoError(t, err)

	assert.Equal(t, numInvocations, 1)
}

func TestPostingTakeActionResult_shouldProcessRequestThroughHandler(t *testing.T) {
	numInvocations := 0
	cleanupFunc := mockYamlHandler(
		func(http.ResponseWriter, *http.Request) { numInvocations++ },
	)
	defer cleanupFunc()

	server := httptest.NewServer(Handler())
	defer server.Close()

	_, err := http.DefaultClient.Post(server.URL+flytepath.TakeActionResultPath, "any content type", nil)
	require.NoError(t, err)

	assert.Equal(t, numInvocations, 1)
}

func TestPostingFlow_shouldProcessRequestThroughHandler(t *testing.T) {
	numInvocations := 0
	cleanupFunc := mockYamlHandler(
		func(http.ResponseWriter, *http.Request) { numInvocations++ },
	)
	defer cleanupFunc()

	server := httptest.NewServer(Handler())
	defer server.Close()

	_, err := http.DefaultClient.Post(server.URL+flytepath.FlowsPath, "any content type", nil)
	require.NoError(t, err)

	assert.Equal(t, numInvocations, 1)
}

func TestPuttingDatastore_shouldProcessRequestThroughHandler(t *testing.T) {
	numInvocations := 0
	cleanupFunc := mockYamlHandler(
		func(http.ResponseWriter, *http.Request) { numInvocations++ },
	)
	defer cleanupFunc()

	server := httptest.NewServer(Handler())
	defer server.Close()

	req, err := http.NewRequest(http.MethodPut, server.URL+flytepath.DatastoreItemPath, nil)
	require.NoError(t, err)
	_, err = http.DefaultClient.Do(req)
	require.NoError(t, err)

	assert.Equal(t, numInvocations, 1)
}

func mockYamlHandler(mockFunc func(http.ResponseWriter, *http.Request)) (cleanupFunc func()) {
	originalYamlHandler := yamlHandler
	yamlHandler = func(http.HandlerFunc) http.HandlerFunc { return mockFunc }
	return func() { yamlHandler = originalYamlHandler }
}
