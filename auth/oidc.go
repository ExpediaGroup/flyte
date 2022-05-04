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
	"errors"
	"fmt"
	"github.com/HotelsDotCom/go-logger"
	"github.com/coreos/go-oidc"
	"github.com/golang-jwt/jwt"
	"github.com/golang-jwt/jwt/request"
	"github.com/husobee/vestigo"
	"net/http"
	"time"
)

func NewAuthHandler(h http.Handler, issuerURL, clientID, policyPath string) (http.Handler, error) {

	pathPolicies, err := newPathPolicies(policyPath)
	if err != nil {
		return nil, err
	}

	verifier, err := createVerifier(issuerURL, clientID)
	if err != nil {
		return nil, err
	}

	// if a path isn't defined in the auth policy then we want to return a 401
	// - this catch all ensures this, rather than vestigo's default behaviour of 404
	router := vestigo.NewRouter()
	router.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))

	for _, p := range pathPolicies {
		for _, m := range p.HttpMethods {
			switch m {
			case http.MethodConnect,
				http.MethodDelete,
				http.MethodGet,
				http.MethodHead,
				http.MethodOptions,
				http.MethodPatch,
				http.MethodPost,
				http.MethodPut,
				http.MethodTrace:
				router.Add(m, p.Path, authHandlerFunc(h, p.Claims, verifier))
			default:
				return nil, errors.New(fmt.Sprintf("Http method %q defined in auth policy file is not supported", m))
			}
		}
		// if no methods specified for path then we want to handle all valid http methods
		if p.HttpMethods == nil {
			router.Handle(p.Path, authHandlerFunc(h, p.Claims, verifier))
		}
	}
	return router, nil
}

var createVerifier = func(issuerURL, clientID string) (*oidc.IDTokenVerifier, error) {
	httpClient := &http.Client{
		Timeout: time.Second * 10,
	}
	provider, err := oidc.NewProvider(oidc.ClientContext(context.Background(), httpClient), issuerURL)
	if err != nil {
		return nil, err
	}

	return provider.Verifier(&oidc.Config{ClientID: clientID}), nil
}

func authHandlerFunc(h http.Handler, c policyClaims, v *oidc.IDTokenVerifier) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if len(c) > 0 {

			token, err := request.AuthorizationHeaderExtractor.ExtractToken(req)
			if err != nil {
				logger.Infof("could not authorize: %v", err)
				w.Header().Set("WWW-Authenticate", `Bearer realm="token"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			idToken, err := v.Verify(req.Context(), token)
			if err != nil {
				logger.Infof("could not authorize: %v", err)
				w.Header().Set("WWW-Authenticate", `Bearer realm="token", error="invalid_token"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			var claims jwt.MapClaims
			if err := idToken.Claims(&claims); err != nil {
				logger.Infof("failed to parse claims: %v", err)
				w.Header().Set("WWW-Authenticate", `Bearer realm="token", error="invalid_token"`)
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if !c.fulfilled(claims, getPathParams(req)) {
				logger.Info("token claims do not satisfy required claims")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
		}
		h.ServeHTTP(w, req)
	}
}

func getPathParams(req *http.Request) map[string]string {
	pathParams := make(map[string]string)
	for _, name := range vestigo.TrimmedParamNames(req) {
		pathParams[name] = vestigo.Param(req, name)
	}
	return pathParams
}
