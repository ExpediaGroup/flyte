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
	"time"
)

const (
	DbName                = "flyte"
	PackCollectionId      = "pack"
	FlowCollectionId      = "flow"
	HistoryCollectionId   = "flowHistory"
	ActionCollectionId    = "action"
	DatastoreCollectionId = "datastore"
)

var (
	session *mgo.Session
	mongoDialTimeout      = 5 * time.Second
	mongoDialRetryWait    = 30 * time.Second
)

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

	_, err := mgo.ParseURL(url)

	if err != nil {
		logger.Fatalf("Invalid mongo url=%s: %v", url, err)
	}

	session = dial(url)

	EnsureIndexExists(ActionCollectionId, "actionCorrelationId", []string{"correlationId"})
	EnsureIndexExists(ActionCollectionId, "actionCompound", []string{"packName", "state.value", "name", "state.time"})
	EnsureTTLIndexExists(ActionCollectionId, "actionTTL", []string{"state.time"}, ttl)
}

func dial(url string) *mgo.Session {

	s, err := mgo.DialWithTimeout(url, mongoDialTimeout)
	if err != nil {
		logger.Errorf("Unable to connect to mongo on url=%s will retry in %s: %v", url, mongoDialRetryWait.String(), err)
		time.Sleep(mongoDialRetryWait)
		return dial(url)
	}
	s.SetSyncTimeout(1 * time.Minute)
	s.SetSocketTimeout(1 * time.Minute)

	return s
}
