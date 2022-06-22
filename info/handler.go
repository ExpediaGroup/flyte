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

package info

import (
	"github.com/ExpediaGroup/flyte/flytepath"
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/ExpediaGroup/flyte/mongo"
	"github.com/rs/zerolog/log"
	"io/ioutil"
	"net/http"
)

var swaggerFileLocation = "swagger/v1.yml"

type Response struct {
	Links []httputil.Link `json:"links"`
}

func Index(w http.ResponseWriter, r *http.Request) {

	links := []httputil.Link{
		{Href: httputil.UriBuilder(r).Path("").Build(), Rel: "self"},
		{Href: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.InfoDoc)).Build(), Rel: "help"},
		{Href: httputil.UriBuilder(r).Path(flytepath.VersionPath).Build(), Rel: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.VersionInfoDoc)).Build()},
	}
	httputil.WriteResponse(w, r, Response{Links: links})
}

func V1(w http.ResponseWriter, r *http.Request) {
	links := []httputil.Link{
		{Href: httputil.UriBuilder(r).Path(flytepath.VersionPath).Build(), Rel: "self"},
		{Href: httputil.UriBuilder(r).Path(flytepath.VersionPath).Parent().Build(), Rel: "up"},
		{Href: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.InfoVersionDoc)).Build(), Rel: "help"},
		{Href: httputil.UriBuilder(r).Path(flytepath.HealthPath).Build(), Rel: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.HealthDoc)).Build()},
		{Href: httputil.UriBuilder(r).Path(flytepath.PacksPath).Build(), Rel: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.ListPacksDoc)).Build()},
		{Href: httputil.UriBuilder(r).Path(flytepath.FlowsPath).Build(), Rel: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.ListFlowDoc)).Build()},
		{Href: httputil.UriBuilder(r).Path(flytepath.DatastorePath).Build(), Rel: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.ListDataItemsDoc)).Build()},
		{Href: httputil.UriBuilder(r).Path(flytepath.AuditFlowPath).Build(), Rel: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.AuditFlowsDoc)).Build()},
		{Href: httputil.UriBuilder(r).Path(flytepath.VersionDocPath).Build(), Rel: httputil.UriBuilder(r).Path(flytepath.GetUriDocPathFor(flytepath.SwaggerRootDoc)).Build()},
	}
	httputil.WriteResponse(w, r, Response{Links: links})
}

func V1Swagger(w http.ResponseWriter, _ *http.Request) {

	swaggerFile, err := ioutil.ReadFile(swaggerFileLocation)
	if err != nil {
		log.Err(err).Msg("cannot read swagger/v1.yml")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(httputil.HeaderContentType, "application/vnd.yaml; charset=utf-8")
	w.Write(swaggerFile)
}

func Health(w http.ResponseWriter, _ *http.Request) {

	if err := mongo.Health(); err != nil {
		log.Err(err).Msg("failed health request")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
