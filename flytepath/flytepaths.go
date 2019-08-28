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

package flytepath

const (
	VersionPath = "/v1"

	// audit
	AuditFlowPath = VersionPath + "/audit/flows"
	AuditGetFlow  = VersionPath + "/audit/flows/:correlationId"
	AuditDoc      = "auditDoc"
	AuditFlowsDoc = "auditFlowsDoc"

	// datastore
	DatastorePath     = VersionPath + "/datastore"
	DatastoreItemPath = VersionPath + "/datastore/:key"

	DatastoreDoc = "datastore"

	// flow
	FlowsPath           = VersionPath + "/flows"
	FlowPath            = VersionPath + "/flows/:flowName"
	FlowExecutionDoc    = "flowExecution"
	TakeActionResultDoc = "takeActionResult"

	FlowDoc = "FlowDoc"

	// info
	HealthPath     = "/health"
	IndexPath      = "/"
	VersionDocPath = VersionPath + "/swagger"

	HealthDoc             = "health"
	InfoDoc               = "infoDoc"
	InfoVersionDoc        = "infoVersionDoc"
	ListDataItemsDoc      = "listDataItems"
	ListFlowDoc           = "listFlows"
	ListFlowExecutionsDoc = "listFlowExecutions"
	ListPacksDoc          = "listPacks"
	SwaggerRootDoc        = "swaggerRoot"
	VersionInfoDoc        = "versionInfoDoc"

	// packs
	PacksPath                 = VersionPath + "/packs"
	PackPath                  = VersionPath + "/packs/:packId"
	PostEventPath             = PackPath + "/events"
	TakeActionPath            = PackPath + "/actions/take"
	TakeActionWithCommandPath = PackPath + "/actions/take?commandName=:commandName"
	TakeActionResultPath      = PackPath + "/actions/:actionId/result"

	GetPacksDoc   = "GetPacksDoc"
	PostEventDoc  = "PostEventDoc"
	TakeActionDoc = "TakeActionDoc"
)

var FlyteDocPaths = map[string]string{
	AuditDoc:              "/swagger#/flowExecs",
	AuditFlowsDoc:         "/swagger#!/flowAudit/findFlows",
	DatastoreDoc:          "/swagger#/datastore",
	FlowExecutionDoc:      "/swagger#!/flow-executions",
	FlowDoc:               "/swagger#/flow",
	InfoDoc:               "/swagger#/info",
	InfoVersionDoc:        "/swagger#!/info" + VersionPath,
	GetPacksDoc:           "/swagger#/pack",
	HealthDoc:             "/swagger#!/info/health",
	ListDataItemsDoc:      "/swagger#!/datastore/listDatastoreItems",
	ListFlowDoc:           "/swagger#!/flow/listFlows",
	ListFlowExecutionsDoc: "/swagger#!/flowExecutions",
	ListPacksDoc:          "/swagger#!/pack/listPacks",
	PostEventDoc:          "/swagger#/event",
	VersionDocPath:        VersionPath + "/swagger",
	SwaggerRootDoc:        "/swagger",
	TakeActionDoc:         "/swagger#!/action/takeAction",
	TakeActionResultDoc:   "/swagger#/actionResult",
	VersionInfoDoc:        "/swagger#!/info" + VersionPath,
}