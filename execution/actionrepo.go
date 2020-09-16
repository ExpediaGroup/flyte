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
	"github.com/ExpediaGroup/flyte/collections"
	"github.com/ExpediaGroup/flyte/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type actionMgoRepo struct{}

func (actionMgoRepo) Add(action Action) error {

	s := mongo.GetSession()
	defer s.Close()

	return s.DB(mongo.DbName).C(mongo.ActionCollectionId).Insert(action)
}

func (actionMgoRepo) FindCorrelated(correlationId string) ([]Action, error) {

	s := mongo.GetSession()
	defer s.Close()

	var actions []Action
	return actions, s.DB(mongo.DbName).
		C(mongo.ActionCollectionId).
		Find(bson.M{"correlationId": correlationId}).
		Select(bson.M{"_id": 1, "stepId": 1, "state": 1}).
		All(&actions)
}

func (actionMgoRepo) FindNew(pack Pack, name string) (*Action, error) {

	s := mongo.GetSession()
	defer s.Close()

	query := bson.M{"packName": pack.Name, "state.value": stateNew}
	if name != "" {
		query["name"] = name
	}

	var actions []Action
	err := s.DB(mongo.DbName).
		C(mongo.ActionCollectionId).
		Find(query).
		Sort("state.time").
		All(&actions)
	if err == mgo.ErrNotFound {
		return nil, nil
	}

	for _, a := range actions {
		if collections.ContainsAll(pack.Labels, a.PackLabels) {
			return &a, nil
		}
	}

	return nil, nil
}

func (actionMgoRepo) Get(actionId string) (*Action, error) {

	s := mongo.GetSession()
	defer s.Close()

	var action Action
	err := s.DB(mongo.DbName).
		C(mongo.ActionCollectionId).
		Find(bson.M{"_id": actionId}).
		One(&action)
	if err == mgo.ErrNotFound {
		return nil, ActionNotFoundErr
	}
	return &action, err
}

func (actionMgoRepo) Update(action Action) error {

	s := mongo.GetSession()
	defer s.Close()

	return s.DB(mongo.DbName).C(mongo.ActionCollectionId).
		Update(bson.M{"_id": action.Id, "state.value": action.prevState.Value}, action)
}
