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
	"github.com/ExpediaGroup/flyte/mongo"
	"github.com/HotelsDotCom/go-logger"
	"gopkg.in/mgo.v2/bson"
)

type flowMgoRepo struct{}

func (r flowMgoRepo) Find(filter flowsFilter) ([]Flow, error) {

	ids, err := findCorrelationIds(filter)
	if err != nil {
		return nil, err
	}

	actions, err := findActionsByCorrelationIds(ids)
	if err != nil {
		return nil, err
	}

	flowsMap := groupActionsIntoFlows(actions)

	return sortFlows(ids, flowsMap), nil
}

func (r flowMgoRepo) Get(correlationId string) (*Flow, error) {

	actions, err := findActionsByCorrelationIds([]string{correlationId})
	if err != nil {
		return nil, err
	}

	flowsMap := groupActionsIntoFlows(actions)

	flow, ok := flowsMap[correlationId]
	if !ok {
		return nil, nil
	}

	return &flow, nil
}

func findCorrelationIds(filter flowsFilter) ([]string, error) {

	s := mongo.GetSession()
	defer s.Close()

	var bsonIds []bson.M
	pipe := s.DB(mongo.DbName).C(mongo.ActionCollectionId).Pipe([]bson.M{
		{"$match": filter.toQuery()},
		{"$group": bson.M{"_id": "$correlationId", "time": bson.M{"$max": "$state.time"}}},
		{"$sort": bson.M{"time": -1}},
		{"$skip": filter.skip},
		{"$limit": filter.limit},
	})
	if err := pipe.All(&bsonIds); err != nil {
		return nil, err
	}

	var ids []string
	for _, bsonId := range bsonIds {
		ids = append(ids, bsonId["_id"].(string))
	}

	return ids, nil
}

func findActionsByCorrelationIds(correlationIds []string) ([]Action, error) {

	s := mongo.GetSession()
	defer s.Close()

	var actions []Action
	return actions, s.DB(mongo.DbName).
		C(mongo.ActionCollectionId).
		Find(bson.M{"correlationId": bson.M{"$in": correlationIds}}).
		All(&actions)
}

func groupActionsIntoFlows(actions []Action) map[string]Flow {

	flowsMap := map[string]Flow{}
	for _, action := range actions {
		if _, ok := flowsMap[action.CorrelationId]; !ok {
			flow, err := getFlow(action.FlowUUID)
			if err != nil {
				logger.Error(err)
				continue
			}
			flow.Actions = map[string]Action{}
			flow.CorrelationId = action.CorrelationId
			flowsMap[flow.CorrelationId] = *flow
		}
		flowsMap[action.CorrelationId].Actions[action.StepId] = action
	}
	return flowsMap
}

func getFlow(uuid string) (*Flow, error) {

	s := mongo.GetSession()
	defer s.Close()

	var flow Flow
	return &flow, s.DB(mongo.DbName).
		C(mongo.HistoryCollectionId).
		Find(bson.M{"uuid": uuid}).
		One(&flow)
}

func sortFlows(correlationIds []string, flowsMap map[string]Flow) []Flow {

	var flows []Flow
	for _, id := range correlationIds {
		if flow, ok := flowsMap[id]; ok {
			flows = append(flows, flow)
		}
	}
	return flows
}

type flowsFilter struct {
	flowName         string
	stepId           string
	actionName       string
	actionPackName   string
	actionPackLabels map[string]string
	skip             int
	limit            int
}

func (flt flowsFilter) toQuery() bson.M {

	query := bson.M{}
	if flt.flowName != "" {
		query["flowName"] = flt.flowName
	}
	if flt.stepId != "" {
		query["stepId"] = flt.stepId
	}
	if flt.actionName != "" {
		query["name"] = flt.actionName
	}
	if flt.actionPackName != "" {
		query["packName"] = flt.actionPackName
	}
	for k, v := range flt.actionPackLabels {
		query["packLabels."+k] = v
	}
	return query
}
