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
	"github.com/ExpediaGroup/flyte/flytepath"
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
	"strings"
)

func toFlowsResponse(r *http.Request, flows []Flow) flowsResponse {

	fs := []flowResponse{}
	for _, f := range flows {
		link := httputil.UriBuilder(r).Path(flytepath.AuditFlowPath, f.CorrelationId).Build()
		flowResponse := flowResponse{Flow: f, Links: []httputil.Link{{Href: link, Rel: "self"}}}
		fs = append(fs, flowResponse)
	}

	return flowsResponse{
		Flows: fs,
		Links: []httputil.Link{
			{Href: httputil.UriBuilder(r).Path(flytepath.AuditFlowPath).Build(), Rel: "self"},
			{Href: httputil.UriBuilder(r).Path(flytepath.AuditFlowPath).Parent().Parent().Build(), Rel: "up"},
			{Href: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.AuditDoc)).Build(), Rel: "help"},
		},
	}
}

func toFlowResponse(r *http.Request, flow Flow) flowResponse {
	link := httputil.UriBuilder(r).Path(flytepath.AuditFlowPath, flow.CorrelationId).Build()
	return flowResponse{Flow: flow, Links: []httputil.Link{{Href: link, Rel: "self"}}}
}

type flowsResponse struct {
	Flows []flowResponse  `json:"flows"`
	Links []httputil.Link `json:"links"`
}

type flowResponse struct {
	Flow
	Links []httputil.Link `json:"links"`
}

func toFlowsFilter(r *http.Request) flowsFilter {

	start := 0
	s := r.URL.Query().Get("start")
	if s != "" {
		if i, err := strconv.Atoi(s); err != nil {
			log.Err(err).Send()
		} else {
			start = i
		}
	}

	limit := 50
	l := r.URL.Query().Get("limit")
	if l != "" {
		if i, err := strconv.Atoi(l); err != nil {
			log.Err(err).Send()
		} else {
			limit = i
		}
	}

	return flowsFilter{
		flowName:         r.URL.Query().Get("flowName"),
		stepId:           r.URL.Query().Get("stepId"),
		actionName:       r.URL.Query().Get("actionName"),
		actionPackName:   r.URL.Query().Get("actionPackName"),
		actionPackLabels: keyValuePair(r.URL.Query().Get("actionPackLabels")),
		skip:             start,
		limit:            limit,
	}
}

// expected format -> env:staging,foo:bar
func keyValuePair(s string) map[string]string {
	if s == "" {
		return map[string]string{}
	}
	result := map[string]string{}
	for _, kvp := range strings.Split(s, ",") {
		if kv := strings.Split(kvp, ":"); len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return result
}
