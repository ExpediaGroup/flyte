// +build acceptance

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

package acceptancetest

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

var AuthFeatures = []Test{
	{"RequestsThatSatisfyPolicyClaimsShouldPassAuthorization", RequestsThatSatisfyPolicyClaimsShouldPassAuthorization},
	{"RequestsThatDoNotSatisfyPolicyClaimsShouldFailAuthorization", RequestsThatDoNotSatisfyPolicyClaimsShouldFailAuthorization},
}

func RequestsThatSatisfyPolicyClaimsShouldPassAuthorization(t *testing.T) {
	// Requires claim "groups":"dev" which token satisfies
	req, err := http.NewRequest(http.MethodGet, flyteApi.SwaggerURL(), nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", dex.GenerateAuthenticIdToken(t)))

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func RequestsThatDoNotSatisfyPolicyClaimsShouldFailAuthorization(t *testing.T) {
	// Requires claim "role":"superuser" which token does not satisfy
	req, err := http.NewRequest(http.MethodPut, flyteApi.SwaggerURL(), nil)
	require.NoError(t, err)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", dex.GenerateAuthenticIdToken(t)))

	resp, err := httpClient.Do(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
