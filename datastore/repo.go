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

package datastore

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/HotelsDotCom/flyte/mongo"
)

type datastoreMgoRepo struct{}

func (r datastoreMgoRepo) Add(dataItem DataItem) error {

	s := mongo.GetSession()
	defer s.Close()

	err := s.DB(mongo.DbName).C(mongo.DatastoreCollectionId).Insert(dataItem)
	if mgo.IsDup(err) {
		err = dataItemExists
	}
	return err
}

func (r datastoreMgoRepo) Remove(key string) error {

	s := mongo.GetSession()
	defer s.Close()

	err := s.DB(mongo.DbName).C(mongo.DatastoreCollectionId).RemoveId(key)
	if err == mgo.ErrNotFound {
		return dataItemNotFound
	}
	return err
}

func (r datastoreMgoRepo) Get(key string) (*DataItem, error) {

	s := mongo.GetSession()
	defer s.Close()

	var dataItem DataItem
	err := s.DB(mongo.DbName).C(mongo.DatastoreCollectionId).FindId(key).One(&dataItem)
	if err == mgo.ErrNotFound {
		return nil, dataItemNotFound
	}
	return &dataItem, err
}

func (r datastoreMgoRepo) FindAll() ([]DataItem, error) {

	s := mongo.GetSession()
	defer s.Close()

	var dataItems []DataItem
	err := s.DB(mongo.DbName).C(mongo.DatastoreCollectionId).Find(nil).
		Select(bson.M{"_id": 1, "description": 1, "contentType": 1}).
		Sort("key").
		All(&dataItems)

	return dataItems, err
}

func (r datastoreMgoRepo) Has(key string) (bool, error){
	//TODO add implementation
	return false, nil
}
