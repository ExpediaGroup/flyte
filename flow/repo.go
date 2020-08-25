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
	"fmt"
	"github.com/ExpediaGroup/flyte/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type flowMgoRepo struct{}

func (r flowMgoRepo) Add(flow Flow) error {

	s := mongo.GetSession()
	defer s.Close()

	if flow.UUID == "" {
		flow.UUID = bson.NewObjectId().Hex()
	}

	if err := s.DB(mongo.DbName).C(mongo.HistoryCollectionId).Insert(flow); err != nil {
		return fmt.Errorf("cannot add to history flow=%+v: %v", flow, err)
	}

	_, err := s.DB(mongo.DbName).C(mongo.FlowCollectionId).Upsert(bson.M{"name": flow.Name}, flow)
	return err
}

func (r flowMgoRepo) Remove(name string) error {

	s := mongo.GetSession()
	defer s.Close()

	err := s.DB(mongo.DbName).C(mongo.FlowCollectionId).Remove(bson.M{"name": name})
	if err == mgo.ErrNotFound {
		err = FlowNotFoundErr
	}
	return err
}

func (r flowMgoRepo) Get(name string) (*Flow, error) {

	s := mongo.GetSession()
	defer s.Close()

	var flow Flow
	err := s.DB(mongo.DbName).C(mongo.FlowCollectionId).Find(bson.M{"name": name}).One(&flow)
	if err == mgo.ErrNotFound {
		return nil, FlowNotFoundErr
	}
	return &flow, err
}

func (r flowMgoRepo) FindAll() ([]Flow, error) {

	s := mongo.GetSession()
	defer s.Close()

	var flows []Flow
	err := s.DB(mongo.DbName).
		C(mongo.FlowCollectionId).
		Find(nil).
		Select(bson.M{"_id": 0, "name": 1, "description": 1}).
		Sort("name").
		All(&flows)

	return flows, err
}
