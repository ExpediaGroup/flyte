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
	"github.com/HotelsDotCom/go-logger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"sync"
	"time"
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

var flowsCache map[string][]Flow

func PurgeFlowsCache(){
	flowsCache = nil
}

func refreshFlowsCache() {
	s := mongo.GetSession()
	defer s.Close()

	flows := []Flow{}
	if err := s.DB(mongo.DbName).C(mongo.FlowCollectionId).Find(bson.D{}).All(&flows); err != nil {
		logger.Errorf("Error refreshing flows cache: %v", err)
		return
	}

	cache := map[string][]Flow{}
	for _, flow := range flows {
		for _, s := range flow.Steps {
			if len(s.DependsOn) > 0 {
				continue
			}
			k := s.Event.PackName + "|" + s.Event.Name
			cache[k] = append(cache[k], flow)
		}
	}
	flowsCache = cache
	logger.Infof("refreshed flow cache; flows in the cache: %d", len(flows))
	return
}

var mu sync.Mutex

func initFlowsCache() {
	mu.Lock()
	defer mu.Unlock()
	if flowsCache != nil {
		return
	}
	refreshFlowsCache()
	go func() {
		for {
			time.Sleep(1 * time.Minute)
			refreshFlowsCache()
		}
	}()
}

func (r flowMgoRepo) FindByEvent(e Event) ([]Flow, error) {
	if flowsCache == nil {
		initFlowsCache()
	}
	k := e.Pack.Name + "|" + e.Name
	flows := []Flow{}
	for _, f := range flowsCache[k] {
		f.correlationId = bson.NewObjectId().Hex()
		f.context = map[string]string{}
		f.actions = map[string]Action{}
		flows = append(flows, f)
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
