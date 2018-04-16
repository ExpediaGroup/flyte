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

package flow

import (
	"net/http"
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/HotelsDotCom/flyte/httputil"
)

type flowResponse struct {
	Flow
	Links []httputil.Link `json:"links"`
}

func toFlowResponse(r *http.Request, flow Flow) flowResponse {

	defaultLinks := []httputil.Link{
		{Href: httputil.UriBuilder(r).Path(flytepath.FlowPath).Replace(":flowName", flow.Name).Build(), Rel: "self"},
		{Href: httputil.UriBuilder(r).Path(flytepath.FlowPath).Parent().Build(), Rel: "up"},
		{Href: flytepath.GetUriDocPathFor(flytepath.FlowDoc), Rel: "help"},
	}
	return flowResponse{
		Flow:  flow,
		Links: defaultLinks,
	}
}

type flowsResponse struct {
	Flows []flowResponse  `json:"flows"`
	Links []httputil.Link `json:"links"`
}

func toFlowsResponse(r *http.Request, flows []Flow) flowsResponse {

	fs := []flowResponse{}
	for _, f := range flows {
		link := httputil.Link{Href: httputil.UriBuilder(r).Path(flytepath.FlowsPath, f.Name).Build(), Rel: "self"}
		fs = append(fs, flowResponse{Flow: f, Links: []httputil.Link{link}})
	}

	defaultLinks := []httputil.Link{
		{Href: httputil.UriBuilder(r).Path(flytepath.FlowsPath).Build(), Rel: "self"},
		{Href: httputil.UriBuilder(r).Path(flytepath.FlowsPath).Parent().Build(), Rel: "up"},
		{Href: flytepath.GetUriDocPathFor(flytepath.FlowDoc), Rel: "help"},
	}
	return flowsResponse{
		Flows: fs,
		Links: defaultLinks,
	}
}
