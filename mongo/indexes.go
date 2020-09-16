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

package mongo

import (
	"github.com/HotelsDotCom/go-logger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
)

func EnsureIndexExists(collectionId, indexName string, indexKey []string) {
	ensureIndex(collectionId, indexName, indexKey, 0)
}

func ensureIndex(collection, indexName string, indexKey []string, ttl int) {
	index := mgo.Index{
		Name:       indexName,
		Key:        indexKey,
		Background: true,
	}
	if ttl > 0 {
		index.ExpireAfter = time.Duration(ttl) * time.Second
	}

	err := ensure(collection, index)
	if err != nil {
		logger.Errorf("Error ensuring index: '%+v', collection: '%s', error: '%v'", indexKey, collection, err)
	}
}

var ensure = ensureIndexFn

func ensureIndexFn(collection string, index mgo.Index) error {
	s := GetSession()
	defer s.Close()

	return s.DB(DbName).C(collection).EnsureIndex(index)
}

func EnsureTTLIndexExists(collectionId, indexName string, indexKey []string, ttl int) {
	if indexExists, index := indexExists(collectionId, indexName); indexExists {
		if indexTTLHasChanged(index.ExpireAfter, ttl) {
			updateTTL(collectionId, indexName, ttl)
		}
	} else {
		ensureIndex(collectionId, indexName, indexKey, ttl)
	}
}

func indexExists(collectionId, indexName string) (bool, mgo.Index) {
	indexes, err := getIndexes(collectionId)
	if err != nil {
		logger.Errorf("Error getting indexes for '%s' collection. Error: '%v'", collectionId, err)
		return false, mgo.Index{}
	}
	for _, i := range indexes {
		if i.Name == indexName {
			return true, i
		}
	}
	return false, mgo.Index{}
}

var getIndexes = getIndexesFn

func getIndexesFn(collectionId string) (indexes []mgo.Index, err error) {
	s := GetSession()
	defer s.Close()

	return s.DB(DbName).C(collectionId).Indexes()
}

func indexTTLHasChanged(currentTTL time.Duration, ttl int) bool {
	if currentTTL != time.Duration(ttl)*time.Second {
		return true
	}
	return false
}

func updateTTL(collection, indexName string, ttl int) {
	if err := update(collection, indexName, ttl); err != nil {
		logger.Errorf("Error updating TTL for '%s' index. TTL: '%v'. Error: '%+v", indexName, ttl, err)
	}
}

var update = updateTTLFn

func updateTTLFn(collection string, indexName string, ttl int) error {
	s := GetSession()
	defer s.Close()

	return s.DB(DbName).Run(bson.D{{"collMod", collection},
		{"index", bson.M{"name": indexName, "expireAfterSeconds": ttl}}}, nil)
}
