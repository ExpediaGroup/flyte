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
	"github.com/husobee/vestigo"
	"net/http"
	"github.com/HotelsDotCom/flyte/audit"
	"github.com/HotelsDotCom/flyte/datastore"
	"github.com/HotelsDotCom/flyte/execution"
	"github.com/HotelsDotCom/flyte/flow"
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/flyte/info"
	"github.com/HotelsDotCom/flyte/pack"
)

func Handler() http.Handler {

	router := vestigo.NewRouter()

	// --- swagger ---
	swaggerUi := http.FileServer(http.Dir("swagger/swagger-ui"))
	router.Handle("/swagger", http.StripPrefix("/swagger", swaggerUi))
	router.Handle("/swagger/:dir/:file", http.StripPrefix("/swagger", swaggerUi))

	// --- info ---
	router.Get(flytepath.IndexPath, info.Index)
	router.Get(flytepath.VersionPath, info.V1)
	router.Get(flytepath.HealthPath, info.Health)
	router.Get(flytepath.VersionDocPath, info.V1Swagger)

	// --- pack ---
	router.Get(flytepath.PacksPath, pack.GetPacks)
	router.Post(flytepath.PacksPath, pack.PostPack, YamlHandler)
	router.Get(flytepath.PackPath, pack.GetPack)
	router.Delete(flytepath.PackPath, pack.DeletePack)

	// --- execution ---
	router.Post(flytepath.TakeActionPath, execution.TakeAction, YamlHandler)
	router.Post(flytepath.PostEventPath, execution.PostEvent, YamlHandler)
	router.Post(flytepath.TakeActionResultPath, execution.CompleteAction, YamlHandler)

	// --- flow ---
	router.Get(flytepath.FlowsPath, flow.GetFlows)
	router.Post(flytepath.FlowsPath, flow.PostFlow, YamlHandler)
	router.Get(flytepath.FlowPath, flow.GetFlow)
	router.Delete(flytepath.FlowPath, flow.DeleteFlow)

	// --- datastore ---
	router.Get(flytepath.DatastorePath, datastore.GetItems)
	router.Get(flytepath.DatastoreItemPath, datastore.GetItem)
	router.Put(flytepath.DatastoreItemPath, datastore.PutItem, YamlHandler)
	router.Delete(flytepath.DatastoreItemPath, datastore.DeleteItem)

	// --- audit ---
	router.Get(flytepath.AuditFlowPath, audit.GetFlows)
	router.Get(flytepath.AuditGetFlow, audit.GetFlow)

	return wrapRequestInterceptorAround(router)
}

func wrapRequestInterceptorAround(h http.Handler) http.Handler {
	return httputil.NewRequestInterceptor(h)
}
