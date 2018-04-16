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

	FlowDoc = "flow"

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

	GetPacksDoc   = "getPacksDoc"
	PostEventDoc  = "postEvent"
	TakeActionDoc = "takeAction"
)

func getFlyteDocPaths() map[string]string {
	relMap := make(map[string]string)
	relMap[AuditDoc] = "/swagger#/flowExecs"
	relMap[AuditFlowsDoc] = "/swagger#!/audit/findFlows"
	relMap[DatastoreDoc] = "/swagger#/datastore"
	relMap[FlowExecutionDoc] = "/swagger#!/flow-executions"
	relMap[FlowDoc] = "/swagger#/flow"
	relMap[InfoDoc] = "/swagger#/info"
	relMap[InfoVersionDoc] = "/swagger#!/info" + VersionPath
	relMap[GetPacksDoc] = "/swagger#/pack"
	relMap[HealthDoc] = "/swagger#!/info/health"
	relMap[ListDataItemsDoc] = "/swagger#!/datastore/listDataItems"
	relMap[ListFlowDoc] = "/swagger#!/flow/listFlows"
	relMap[ListFlowExecutionsDoc] = "/swagger#!/flowExecutions"
	relMap[ListPacksDoc] = "/swagger#!/pack/listPacks"
	relMap[PostEventDoc] = "/swagger#/event"
	relMap["swagger"] = VersionPath + "/swagger"
	relMap[SwaggerRootDoc] = "/swagger"
	relMap[TakeActionDoc] = "/swagger#!/action/takeAction"
	relMap[TakeActionResultDoc] = "/swagger#/actionResult"
	relMap[VersionInfoDoc] = "/swagger#!/info" + VersionPath
	return relMap
}
