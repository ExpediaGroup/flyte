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
	"encoding/base64"
	"fmt"

	"github.com/HotelsDotCom/go-logger"

	"github.com/golang-jwt/jwt"

	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/husobee/vestigo"
	"github.com/stretchr/testify/require"
)

const privateKeyPath = "./testdata/private.pem"

type MockDex struct {
	server *httptest.Server
	key    MockDexKey
}

type MockDexKey struct {
	n string
	e string
}

func StartDex() *MockDex {

	dex := &MockDex{}
	if err := dex.key.initialise(); err != nil {
		logger.Fatalf("Unable to start mock dex: %v", err)
	}

	router := vestigo.NewRouter()
	router.Get("/dex/.well-known/openid-configuration", dex.oidcConfigurationHandler)
	router.Get("/dex/keys", dex.keysHandler)

	dex.server = httptest.NewServer(router)
	return dex
}

func (m *MockDexKey) initialise() error {
	privateKey, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return err
	}

	parsedPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return err
	}

	m.n = base64.RawURLEncoding.EncodeToString(parsedPrivateKey.N.Bytes())
	m.e = base64.RawURLEncoding.EncodeToString(big.NewInt(int64(parsedPrivateKey.E)).Bytes())
	return nil
}

func (m MockDex) IssuerURL() string {
	return m.server.URL + "/dex"
}

func (m MockDex) Stop() {
	if m.server != nil {
		m.server.Close()
	}
}

func (m MockDex) keysHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`{
		"keys": [
			{
				"use": "sig",
				"kty": "RSA",
				"kid": "8aa6f2b108f9b696722451f5b1820a5c452a79a2",
				"alg": "RS256",
				"n": "%s",
				"e": "%s"
			}
		]
	}`, m.key.n, m.key.e)))
}

func (m *MockDex) oidcConfigurationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Write([]byte(fmt.Sprintf(`{
	 "issuer": "%s",
	 "authorization_endpoint": "%s/auth",
	 "token_endpoint": "%s/token",
	 "jwks_uri": "%s/keys",
	 "response_types_supported": ["code"],
	 "subject_types_supported": ["public"],
	 "id_token_signing_alg_values_supported": ["RS256"],
	 "scopes_supported": ["openid","email","groups","profile","offline_access"],
	 "token_endpoint_auth_methods_supported": ["client_secret_basic"],
	 "claims_supported": ["aud","email","email_verified","exp","iat","iss","locale","name","sub"]
	}`, m.IssuerURL(), m.IssuerURL(), m.IssuerURL(), m.IssuerURL())))
}

func (m MockDex) GenerateAuthenticIdToken(t *testing.T) string {
	privateKey, err := ioutil.ReadFile(privateKeyPath)
	require.NoError(t, err, "failed to load private key file")

	parsedPrivateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	require.NoError(t, err, "failed to parse private key data")

	authenticIdToken := jwt.MapClaims{
		"iss":            m.IssuerURL(),
		"sub":            "CggwMDAwMDAwMhIIbW9ja0xEQVA",
		"aud":            "example-app",
		"exp":            4666936924,
		"iat":            1513336924,
		"at_hash":        "T1IfA1zaueyiz1JOoAKArw",
		"email":          "jdoe@email.com",
		"email_verified": true,
		"groups":         []string{"dev"},
		"name":           "Jane Doe",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, authenticIdToken)
	tokenString, err := token.SignedString(parsedPrivateKey)
	require.NoError(t, err)
	return tokenString
}
