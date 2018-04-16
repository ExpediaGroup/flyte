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
	"gopkg.in/mgo.v2"
	"github.com/HotelsDotCom/go-logger"
)

const (
	DbName                = "flyte"
	PackCollectionId      = "pack"
	FlowCollectionId      = "flow"
	HistoryCollectionId   = "flowHistory"
	ActionCollectionId    = "action"
	DatastoreCollectionId = "datastore"
)

var session *mgo.Session

func GetSession() *mgo.Session {
	if session == nil {
		logger.Fatal("Mongo session has not been initialised.")
	}
	return session.Copy()
}

func Health() error {
	s := GetSession()
	defer s.Close()
	return s.Ping()
}

func InitSession(url string, ttl int) {
	s, err := mgo.Dial(url)
	if err != nil {
		logger.Fatalf("Unable to connect to mongo on url=%s: %v", url, err)
	}
	session = s

	EnsureIndexExists(ActionCollectionId, "actionCorrelationId", []string{"correlationId"})
	EnsureIndexExists(ActionCollectionId, "actionCompound", []string{"packName", "state.value", "name", "state.time"})
	EnsureTTLIndexExists(ActionCollectionId, "actionTTL", []string{"state.time"}, ttl)
}
