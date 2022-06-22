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

package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/go-oidc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestShouldReturn200_WhenUserRequestsUnprotectedResourceWithoutNeedForAuthorization(t *testing.T) {

	handler, cleanupFunc := createTestAuthHandler(t, simpleHandler)
	defer cleanupFunc()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://flyte/packs", nil)

	handler.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestShouldReturn401WithHeader_WhenUserRequestsProtectedResourceWithoutIdToken(t *testing.T) {
	handler, cleanupFunc := createTestAuthHandler(t, simpleHandler)
	defer cleanupFunc()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "http://flyte/packs/foo-pack", nil)

	handler.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, `Bearer realm="token"`, resp.Header.Get("WWW-Authenticate"))
}

func TestShouldReturn401WithHeader_WhenUserRequestsProtectedResourceWithInvalidIdToken(t *testing.T) {
	handler, cleanupFunc := createTestAuthHandler(t, simpleHandler)
	defer cleanupFunc()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "http://flyte/packs/foo-pack", nil)
	invalidIdToken := "this-is-not-a-valid-token"
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", invalidIdToken))

	handler.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, `Bearer realm="token", error="invalid_token"`, resp.Header.Get("WWW-Authenticate"))
}

func TestShouldReturn401_WhenUserRequestsProtectedResourceWithExpiredIdToken(t *testing.T) {
	handler, cleanupFunc := createTestAuthHandler(t, simpleHandler)
	defer cleanupFunc()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "http://flyte/packs/foo-pack", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", expiredIdToken))

	handler.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	assert.Equal(t, `Bearer realm="token", error="invalid_token"`, resp.Header.Get("WWW-Authenticate"))
}

func TestShouldReturn401_WhenUserRequestsProtectedResourceWithValidIdTokenButNoMatchingClaims(t *testing.T) {
	handler, cleanupFunc := createTestAuthHandler(t, simpleHandler)
	defer cleanupFunc()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "http://flyte/packs/foo-pack", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authenticIdToken))

	handler.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestShouldReturn200_WhenUserRequestsProtectedResourceWithValidIdTokenAndMatchingClaims(t *testing.T) {

	handler, cleanupFunc := createTestAuthHandler(t, simpleHandler)
	defer cleanupFunc()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "http://flyte/packs/foo-pack", nil)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", authenticIdToken))

	handler.ServeHTTP(w, req)
	resp := w.Result()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// -- mocks, test data and setup functions

var simpleHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "hello!")
})

func createTestAuthHandler(t *testing.T, h http.Handler) (handler http.Handler, cleanupFunc func()) {
	var jwk jose.JSONWebKey
	require.NoError(t, json.Unmarshal([]byte(keySet), &jwk))

	originalCreateVerifier := createVerifier
	createVerifier = func(issuerURL, clientID string) (*oidc.IDTokenVerifier, error) {
		return oidc.NewVerifier(issuerURL, &testKeySet{jwk}, &oidc.Config{ClientID: clientID}), nil
	}

	handler, err := NewAuthHandler(h, "http://127.0.0.1:5556/dex", "example-app", policyPath)
	require.NoError(t, err)

	return handler, func() { createVerifier = originalCreateVerifier }
}

type testKeySet struct {
	jwk jose.JSONWebKey
}

func (t *testKeySet) VerifySignature(ctx context.Context, jwt string) ([]byte, error) {
	jws, err := jose.ParseSigned(jwt)
	if err != nil {
		return nil, fmt.Errorf("oidc: malformed jwt: %v", err)
	}
	return jws.Verify(&t.jwk)
}

const (
	policyPath = "./testdata/policy_config.yaml"

	keySet = `{
	"use": "sig",
	"kty": "RSA",
	"kid": "8aa6f2b108f9b696722451f5b1820a5c452a79a2",
	"alg": "RS256",
	"n": "xJKdq-9PRrxeilv85DB005l76MelgemRr3wXxH8PLqJ8dEq5FfJ-8TVxy7Oq9dQoW2_KmmZnhQBqVhH26YkKsA1AZ-r9kaPb-XVJYKE8eKR5FgkuE9N2pPYB5aYXduXuLt2KjGzh4RzcWPWs5NVeeN5qPBLWhaekTucIIsHLKSq4cgaic9jragwK2WN7coQluMJ6J8KZTCr7ibKEVCvaf9TuAIqO9uG2n2fLeyNjLASflsp5zeAxdJTZHJpvQpUrPyueyn6tYKwo09Vohu2BbQakEe5QzB2RKEBFc8ltSE7CBeLHav7wt5nnHJJRSdKxCck9hpII4TiZ8V9x2AEDow",
	"e": "AQAB"
	}`

	/*
		Claims of expiredIdToken:

		{
			"iss": "http://127.0.0.1:5556/dex",
			"sub": "CggwMDAwMDAwMRIIbW9ja0xEQVA",
			"aud": "example-app",
			"exp": 1513244622,
			"iat": 1513241022,
			"at_hash": "Fb9xmVVOQU9RGWsGb9__Yg",
			"email": "jdoe@email.com",
			"email_verified": true,
			"groups": [
				"packadmin"
			],
			"name": "John Doe"
		}
	*/
	expiredIdToken = `eyJhbGciOiJSUzI1NiIsImtpZCI6IjhhYTZmMmIxMDhmOWI2OTY3MjI0NTFmNWIxODIwYTVjNDUyYTc5YTIifQ.eyJpc3MiOiJodHRwOi8vMTI3LjAuMC4xOjU1NTYvZGV4Iiwic3ViIjoiQ2dnd01EQXdNREF3TVJJSWJXOWphMHhFUVZBIiwiYXVkIjoiZXhhbXBsZS1hcHAiLCJleHAiOjE1MTMyNDQ2MjIsImlhdCI6MTUxMzI0MTAyMiwiYXRfaGFzaCI6IkZiOXhtVlZPUVU5UkdXc0diOV9fWWciLCJlbWFpbCI6Impkb2VAZW1haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImdyb3VwcyI6WyJwYWNrYWRtaW4iXSwibmFtZSI6IkpvaG4gRG9lIn0.TsEUdDcEWKdq9MpuPz9HgxgSF8eUul8RxgzG71NOXJMQTibSRYDFGK7QHYMTwFSgYLCIpLrn9kRiSvL-bzMc3BlIpCFJiNUm__Q3ZHsnzoNFyc1jqxzgTpAenUBUeRPVi-7q4YsIUhzCuH93UbK9huzOU5z804pXKpzU3zeL481orNiIJs_ZGGkF3lUKscSoaFyN0fVVodxvX1TJJf-r-PqiQclgtNU-dn3rmfhm5mRDAb6k4P_g91YAxDhtpRFOc_UuGQQWaGxeYbQQcBHAYnDIOpUrYqEdbf-pf5ipnNEaDjmZGt6YRKbz3Fm7aLz_1MCWTo7Qec0WneNbGxaWgg`

	/*
		Claims of authenticIdToken:
		{
			"iss": "http://127.0.0.1:5556/dex",
			"sub": "CggwMDAwMDAwMRIIbW9ja0xEQVA",
			"aud": "example-app",
			"exp": 4666840823,
			"iat": 1513240823,
			"at_hash": "p6B8GP9jZ8wbxWd49UGp_A",
			"email": "jdoe@email.com",
			"email_verified": true,
			"groups": [
				"packadmin"
			],
			"name": "John Doe"
		}
	*/
	authenticIdToken = `eyJhbGciOiJSUzI1NiIsImtpZCI6IjhhYTZmMmIxMDhmOWI2OTY3MjI0NTFmNWIxODIwYTVjNDUyYTc5YTIifQ.eyJpc3MiOiJodHRwOi8vMTI3LjAuMC4xOjU1NTYvZGV4Iiwic3ViIjoiQ2dnd01EQXdNREF3TVJJSWJXOWphMHhFUVZBIiwiYXVkIjoiZXhhbXBsZS1hcHAiLCJleHAiOjQ2NjY4NDA4MjMsImlhdCI6MTUxMzI0MDgyMywiYXRfaGFzaCI6InA2QjhHUDlqWjh3YnhXZDQ5VUdwX0EiLCJlbWFpbCI6Impkb2VAZW1haWwuY29tIiwiZW1haWxfdmVyaWZpZWQiOnRydWUsImdyb3VwcyI6WyJwYWNrYWRtaW4iXSwibmFtZSI6IkpvaG4gRG9lIn0.nebCA0fB35bdKFiTM7JVrJR3lH1lMr9K5AzvgC_NVL0d272c94gcpDQHRyJSMVj2C2Ldv0_PCToM8tjJnjfg17uCY6GbXmaYHlvLF3VrL__TL5zoX0PnvTTLEAv-zv1NTer2mZXLVOorr6hKYyFNmitn-kaWv8vjKPMDqrs6hpCUO1Yf3ACdbsJ35qmmm9OkzAGCcHCsNwMMrMh_evsrnDRPYYp19wHuqvPErjgUbJrxBnk-wqcp-KaQtqnZSe1Ajlu0KL1tRVhQYxThYhWtUvDOHQacTm1KisKmUFPnjJiKphb9lbUCfa6cGai14SreFqOcIwyyo2sJV6IF1qj0FQ`
)
