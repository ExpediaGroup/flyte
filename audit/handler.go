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

package audit

import (
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/husobee/vestigo"
	"github.com/rs/zerolog/log"
	"net/http"
)

var flowRepo Repository = flowMgoRepo{}

func GetFlows(w http.ResponseWriter, r *http.Request) {

	flows, err := flowRepo.Find(toFlowsFilter(r))
	if err != nil {
		log.Err(err).Send()
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputil.WriteResponse(w, r, toFlowsResponse(r, flows))
}

func GetFlow(w http.ResponseWriter, r *http.Request) {

	correlationId := vestigo.Param(r, "correlationId")
	flow, err := flowRepo.Get(correlationId)

	if err != nil {
		log.Err(err).Msgf("Error finding flow correlationId=%s", correlationId)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if flow == nil {
		log.Info().Msgf("Flow correlationId=%s not found", correlationId)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	httputil.WriteResponse(w, r, toFlowResponse(r, *flow))
}
