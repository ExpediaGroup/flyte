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

package execution

import (
	"fmt"
	"github.com/ExpediaGroup/flyte/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type flowMgoRepo struct{}

func (r flowMgoRepo) GetByAction(action Action) (*Flow, error) {

	flow, err := r.getFlow(action.FlowUUID)
	if err != nil {
		return flow, err
	}
	if flow == nil {
		return nil, fmt.Errorf("flow with uuid=%s not found", action.FlowUUID)
	}

	actions, err := actionRepo.FindCorrelated(action.CorrelationId)
	if err != nil {
		return flow, err
	}

	flow.correlationId = action.CorrelationId
	flow.context = action.Context
	flow.actions = map[string]Action{}

	for _, a := range actions {
		flow.actions[a.StepId] = a
	}

	return flow, nil
}

func (r flowMgoRepo) FindByEvent(e Event) ([]Flow, error) {

	s := mongo.GetSession()
	defer s.Close()

	flowQuery := bson.M{
		"steps": bson.M{
			"$elemMatch": bson.M{
				"event.packName": e.Pack.Name,
				"event.name":     e.Name,
				"dependsOn": bson.M{
					"$exists": false,
				},
			},
		},
	}

	flows := []Flow{}
	if err := s.DB(mongo.DbName).C(mongo.FlowCollectionId).Find(flowQuery).All(&flows); err != nil {
		return flows, err
	}

	for i := range flows {
		flows[i].correlationId = bson.NewObjectId().Hex()
		flows[i].context = map[string]string{}
		flows[i].actions = map[string]Action{}
	}
	return flows, nil
}

func (r flowMgoRepo) getFlow(uuid string) (*Flow, error) {

	s := mongo.GetSession()
	defer s.Close()

	var flow Flow
	err := s.DB(mongo.DbName).C(mongo.HistoryCollectionId).Find(bson.M{"uuid": uuid}).One(&flow)
	if err == mgo.ErrNotFound {
		return nil, nil
	}
	return &flow, err
}
