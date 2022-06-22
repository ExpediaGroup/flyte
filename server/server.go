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
	"fmt"
	"github.com/ExpediaGroup/flyte/auth"
	"github.com/ExpediaGroup/flyte/mongo"
	"github.com/rs/zerolog/log"
	"net/http"
)

type FlyteServer struct {
	*http.Server
}

func NewFlyteServer(port, mgoHost string, ttl int) *FlyteServer {

	mongo.InitSession(mgoHost, ttl)
	return &FlyteServer{
		&http.Server{
			Addr:    fmt.Sprintf(":%s", port),
			Handler: Handler(),
		},
	}
}

func (f *FlyteServer) EnableAuth(authPolicyPath, oidcIssuerURL, oidcClientID string) {
	authHandler, err := auth.NewAuthHandler(f.Handler, oidcIssuerURL, oidcClientID, authPolicyPath)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to enable auth")
	}

	f.Handler = authHandler
	log.Info().Msgf("Enabled auth using auth policy file %q and OIDC issuer uri %q and OIDC issuer client id %q", authPolicyPath, oidcIssuerURL, oidcClientID)
}
