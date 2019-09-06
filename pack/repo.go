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

package pack

import (
	"github.com/HotelsDotCom/flyte/mongo"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

type packMgoRepo struct{}

func (r packMgoRepo) Add(pack Pack) error {

	s := mongo.GetSession()
	defer s.Close()

	_, err := s.DB(mongo.DbName).C(mongo.PackCollectionId).UpsertId(pack.Id, pack)
	return err
}

func (r packMgoRepo) Remove(id string) error {

	s := mongo.GetSession()
	defer s.Close()

	err := s.DB(mongo.DbName).C(mongo.PackCollectionId).RemoveId(id)
	if err == mgo.ErrNotFound {
		return PackNotFoundErr
	}
	return err
}

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

func (r packMgoRepo) FindAll() ([]Pack, error) {

	s := mongo.GetSession()
	defer s.Close()

	var ps []Pack
	err := s.DB(mongo.DbName).
		C(mongo.PackCollectionId).
		Find(nil).
		Select(bson.M{"_id": 1, "name": 1, "lastSeen": 1, "labels": 1}).
		Sort("name").
		All(&ps)

	return ps, err
}

func (r packMgoRepo) RemoveAllOlderThan(date time.Time) (packsRemoved int, err error) {

	s := mongo.GetSession()
	defer s.Close()

	info, err := s.DB(mongo.DbName).C(mongo.PackCollectionId).RemoveAll(bson.M{"lastSeen": bson.M{"$lt": date}})
	if err != nil {
		return 0, err
	}
	return info.Removed, nil
}