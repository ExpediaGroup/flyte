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

package main

import (
	"github.com/ExpediaGroup/flyte/pack"
	"github.com/ExpediaGroup/flyte/server"
	"github.com/rs/zerolog/log"
	"net/http"
	"os"
)

func main() {
	log.Logger = log.Output(os.Stdout)

	c := NewConfig()

	if c.ShouldDeleteDeadPacks {
		log.Info().Msgf("daily removal of dead packs is scheduled to run at '%s' set with a grace period of '%v' seconds.", c.DeleteDeadPacksTime, c.PackGracePeriodUntilDeadInSeconds)

		pack.ScheduleDailyRemovalOfDeadPacksAt(c.DeleteDeadPacksTime, c.PackGracePeriodUntilDeadInSeconds)
	}

	flyteServer := server.NewFlyteServer(c.Port, c.MongoHost, c.FlyteTTL)

	if c.requireAuth() {
		flyteServer.EnableAuth(c.AuthPolicyPath, c.OidcIssuerURL, c.OidcIssuerClientID)
	}

	log.Info().Msgf("Serving flyteapi on %s with TLS %v", flyteServer.Addr, c.requireTLS())

	var err error
	if c.requireTLS() {
		err = flyteServer.ListenAndServeTLS(c.TLSCertPath, c.TLSKeyPath)
	} else {
		err = flyteServer.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		log.Fatal().Msgf("flyteapi server failure: %s", err)
	}
}
