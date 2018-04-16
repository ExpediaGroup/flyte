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
	"gopkg.in/mgo.v2"
	"github.com/HotelsDotCom/flyte/mongo"
)

type packMgoRepo struct{}

func (r packMgoRepo) Get(id string) (*Pack, error) {

	s := mongo.GetSession()
	defer s.Close()

	var pack Pack
	err := s.DB(mongo.DbName).C(mongo.PackCollectionId).FindId(id).One(&pack)
	if err == mgo.ErrNotFound {
		return nil, PackNotFoundErr
	}
	return &pack, err
}
