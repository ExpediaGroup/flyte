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
	"github.com/ExpediaGroup/flyte/mongo"
	"gopkg.in/mgo.v2/bson"
)

type auditMgoRepo struct{}

func (auditMgoRepo) Add(action Action) error {

	s := mongo.GetSession()
	defer s.Close()

	return s.DB(mongo.DbName).C(mongo.AuditCollectionId).Insert(action)
}

func (auditMgoRepo) Update(action Action) error {

	s := mongo.GetSession()
	defer s.Close()

	return s.DB(mongo.DbName).C(mongo.AuditCollectionId).
		Update(bson.M{"_id": action.Id, "state.value": action.prevState.Value}, action)
}
