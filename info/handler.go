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
	"io/ioutil"
	"net/http"
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/flyte/mongo"
	"github.com/HotelsDotCom/go-logger"
)

var swaggerFileLocation = "swagger/v1.yml"

type Response struct {
	Links []httputil.Link `json:"links"`
}

func Index(w http.ResponseWriter, r *http.Request) {

	links := []httputil.Link{
		{Href: httputil.UriBuilder(r).Path("").Build(), Rel: "self"},
		{Href: flytepath.GetUriDocPathFor(flytepath.InfoDoc), Rel: "help"},
		{Href: httputil.UriBuilder(r).Path(flytepath.VersionPath).Build(), Rel: flytepath.GetUriDocPathFor(flytepath.VersionInfoDoc)},
	}
	httputil.WriteResponse(w, r, Response{Links: links})
}

func V1(w http.ResponseWriter, r *http.Request) {
	links := []httputil.Link{
		{Href: httputil.UriBuilder(r).Path(flytepath.VersionPath).Build(), Rel: "self"},
		{Href: httputil.UriBuilder(r).Path(flytepath.VersionPath).Parent().Build(), Rel: "up"},
		{Href: flytepath.GetUriDocPathFor(flytepath.InfoVersionDoc), Rel: "help"},
		{Href: httputil.UriBuilder(r).Path(flytepath.HealthPath).Build(), Rel: flytepath.GetUriDocPathFor(flytepath.HealthDoc)},
		{Href: httputil.UriBuilder(r).Path(flytepath.PacksPath).Build(), Rel: flytepath.GetUriDocPathFor(flytepath.ListPacksDoc)},
		{Href: httputil.UriBuilder(r).Path(flytepath.FlowsPath).Build(), Rel: flytepath.GetUriDocPathFor(flytepath.ListFlowDoc)},
		{Href: httputil.UriBuilder(r).Path(flytepath.DatastorePath).Build(), Rel: flytepath.GetUriDocPathFor(flytepath.ListDataItemsDoc)},
		{Href: httputil.UriBuilder(r).Path(flytepath.AuditFlowPath).Build(), Rel: flytepath.GetUriDocPathFor(flytepath.AuditFlowsDoc)},
		{Href: httputil.UriBuilder(r).Path(flytepath.VersionDocPath).Build(), Rel: flytepath.GetUriDocPathFor(flytepath.SwaggerRootDoc)},
	}
	httputil.WriteResponse(w, r, Response{Links: links})
}

func V1Swagger(w http.ResponseWriter, _ *http.Request) {

	swaggerFile, err := ioutil.ReadFile(swaggerFileLocation)
	if err != nil {
		logger.Errorf("cannot read swagger/v1.yml: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set(httputil.HeaderContentType, "application/vnd.yaml; charset=utf-8")
	w.Write(swaggerFile)
}

func Health(w http.ResponseWriter, _ *http.Request) {

	if err := mongo.Health(); err != nil {
		logger.Errorf("failed health request: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
