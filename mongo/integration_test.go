// +build slow

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
	"errors"
	"github.com/ExpediaGroup/flyte/mongo/mongotest"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"testing"
	"time"
)

const ttl = 365 * 24 * 60 * 60

var mongoT *mongotest.MongoT

func TestMain(m *testing.M) {
	os.Exit(runTestsWithMongo(m))
}

func runTestsWithMongo(m *testing.M) int {
	mongoT = mongotest.NewMongoT(DbName)
	defer mongoT.Teardown()

	mongoT.Start()

	return m.Run()
}

func Test_InitSession_ShouldCreateIndexesIfTheyDontExist(t *testing.T) {
	// given
	cleanDbPopulatedWithActions(t)
	defaultIndexOnlyExists(t)

	// when
	InitSession(mongoT.GetUrl(), ttl)

	// then...
	indexes, _ := mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	assertIndexesExist(t, indexes, ttl)
}

func Test_InitSession_ShouldNotAffectDbIfIndexesAlreadyExist(t *testing.T) {
	// given
	cleanDbPopulatedWithActions(t)
	defaultIndexOnlyExists(t)

	// and index added
	mongoT.GetSession().DB(DbName).C(ActionCollectionId).EnsureIndex(mgo.Index{
		Key:        []string{"correlationId"},
		Name:       "actionCorrelationId",
		Background: true,
	})
	indexes, _ := mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	require.True(t, len(indexes) == 2)

	// when
	InitSession(mongoT.GetUrl(), ttl)

	// then expected indexes still exist
	indexes, _ = mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	assertIndexesExist(t, indexes, ttl)
}

func Test_InitSession_ShouldLogErrorMessageWhenErrorEnsuringIndexesExist(t *testing.T) {
	// given
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	cleanDbPopulatedWithActions(t)
	defaultIndexOnlyExists(t)

	defer resetEnsureIndexes()
	ensure = func(collection string, index mgo.Index) error {
		return errors.New("some error")
	}

	// when
	InitSession(mongoT.GetUrl(), ttl)

	// then...
	logMessages := loggertest.GetLogMessages()
	assert.Equal(t, "Error ensuring index: '[correlationId]', collection: 'action', error: 'some error'", logMessages[0].Message)
}

func Test_InitSession_ShouldCreateTTLIndexAndExpireRecords(t *testing.T) {
	// given
	cleanDbPopulatedWithActions(t)
	defaultIndexOnlyExists(t)
	ttlInSeconds := 10

	// when
	InitSession(mongoT.GetUrl(), ttlInSeconds)

	// then ensure all is as expected
	indexes, _ := mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	assertIndexesExist(t, indexes, ttlInSeconds)

	// and sleep - set at 61 seconds as by default mongo checks expiry once a minute
	time.Sleep(time.Duration(61) * time.Second)

	// now records should be expired
	assertActionsAreNoLongerInDb(t)
}

func Test_InitSession_ShouldUpdateTTLIndex(t *testing.T) {
	// given
	cleanDbPopulatedWithActions(t)
	defaultIndexOnlyExists(t)
	ttlInSeconds := 10

	InitSession(mongoT.GetUrl(), ttlInSeconds)

	// and all is as expected
	indexes, _ := mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	assertIndexesExist(t, indexes, ttlInSeconds)

	// when data ttl is changed
	newTTLInSeconds := 120
	InitSession(mongoT.GetUrl(), newTTLInSeconds)

	// then check indexes again
	indexes2, _ := mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	assertIndexesExist(t, indexes2, newTTLInSeconds)

	// and sleep - set at 61 seconds as by default mongo checks expiry once a minute
	time.Sleep(time.Duration(61) * time.Second)

	// now records should be not be expired as the TTL has been changed to longer than 60 seconds
	assertActionsAreStillInDb(t)
}

func Test_InitSession_ShouldNotUpdateTTLIndex(t *testing.T) {
	// given
	cleanDbPopulatedWithActions(t)
	defaultIndexOnlyExists(t)
	ttlInSeconds := 10

	InitSession(mongoT.GetUrl(), ttlInSeconds)

	// and all is as expected
	indexes, _ := mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	assertIndexesExist(t, indexes, ttlInSeconds)

	// when data ttl is kept the same...
	InitSession(mongoT.GetUrl(), ttlInSeconds)

	// then indexes should be unchanged
	indexes2, _ := mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	assertIndexesExist(t, indexes2, ttlInSeconds)
}

func Test_InitSession_ShouldLogWhenErrorWhenTryingToUpdateTTLIndex(t *testing.T) {
	// given
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	cleanDbPopulatedWithActions(t)
	defaultIndexOnlyExists(t)
	ttlInSeconds := 1

	InitSession(mongoT.GetUrl(), ttlInSeconds)

	// and all is as expected
	indexes, _ := mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	assertIndexesExist(t, indexes, ttlInSeconds)

	// when updating TTL fails with an error...
	defer resetUpdateTTL()
	update = func(collection string, indexName string, ttl int) error {
		return errors.New("some error")
	}

	newTTLInSeconds := 2
	InitSession(mongoT.GetUrl(), newTTLInSeconds)

	// then
	logMessages := loggertest.GetLogMessages()
	assert.Equal(t, "Error updating TTL for 'actionTTL' index. TTL: '2'. Error: 'some error", logMessages[0].Message)
}

func Test_InitSession_ShouldLogErrorWhenTryingToGetIndexesIfErrorIsReturned(t *testing.T) {
	// given
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	cleanDbPopulatedWithActions(t)
	defaultIndexOnlyExists(t)
	ttlInSeconds := 1

	defer resetGetIndexes()
	getIndexes = func(collectionId string) (indexes []mgo.Index, err error) {
		return []mgo.Index{}, errors.New("some error")
	}

	// when
	InitSession(mongoT.GetUrl(), ttlInSeconds)

	// then
	logMessages := loggertest.GetLogMessages()
	assert.Equal(t, "Error getting indexes for 'action' collection. Error: 'some error'", logMessages[0].Message)
}

func Test_InitSession_ShouldCreateTTLIndexForAuditAndExpireRecords(t *testing.T) {
	// given
	mongoT.DropDatabase(t)
	mongoT.Insert(t, AuditCollectionId, newActionT("1", "actionA", "NEW", time.Now()))
	mongoT.Insert(t, AuditCollectionId, newActionT("2", "actionB", "NEW", time.Now()))
	mongoT.Insert(t, AuditCollectionId, newActionT("3", "actionC", "NEW", time.Now()))

	prevAuditTTL := auditTTL
	defer func() {auditTTL = prevAuditTTL}()
	auditTTL = 10

	// when
	InitSession(mongoT.GetUrl(), 10)
	var got Action
	err := mongoT.FindOne(AuditCollectionId, bson.M{"_id": "1"}, &got)
	assert.NoError(t, err)
	assert.Equal(t, "1", got.Id)

	// then ensure all is as expected
	indexes, _ := mongoT.GetSession().DB(DbName).C(AuditCollectionId).Indexes()
	require.True(t, len(indexes) == 3)
	assert.Contains(t, indexes[0].Key, "_id")
	assert.Contains(t, indexes[1].Key, "correlationId")
	assert.Contains(t, indexes[2].Key, "state.time")
	assert.Equal(t, time.Duration(auditTTL)*time.Second, indexes[2].ExpireAfter)

	// and sleep - set at 61 seconds as by default mongo checks expiry once a minute
	time.Sleep(time.Duration(61) * time.Second)

	// now records should be expired
	err = mongoT.FindOne(AuditCollectionId, bson.M{"_id": "1"}, &got)
	assert.Equal(t, "not found", err.Error())
	err = mongoT.FindOne(AuditCollectionId, bson.M{"_id": "2"}, &got)
	assert.Equal(t, "not found", err.Error())
	err = mongoT.FindOne(AuditCollectionId, bson.M{"_id": "3"}, &got)
	assert.Equal(t, "not found", err.Error())
}

func cleanDbPopulatedWithActions(t *testing.T) {
	mongoT.DropDatabase(t)
	mongoT.Insert(t, ActionCollectionId, newActionT("1", "actionA", "NEW", time.Now()))
	mongoT.Insert(t, ActionCollectionId, newActionT("2", "actionB", "NEW", time.Now()))
	mongoT.Insert(t, ActionCollectionId, newActionT("3", "actionC", "NEW", time.Now()))
}

func defaultIndexOnlyExists(t *testing.T) {
	index, _ := mongoT.GetSession().DB(DbName).C(ActionCollectionId).Indexes()
	require.True(t, len(index) == 1)
	require.True(t, index[0].Key[0] == "_id")
}

func assertIndexesExist(t *testing.T, indexes []mgo.Index, ttl int) {
	require.True(t, len(indexes) == 4)
	assert.Contains(t, indexes[0].Key, "_id")
	assert.Contains(t, indexes[1].Key, "packName")
	assert.Contains(t, indexes[1].Key, "state.value")
	assert.Contains(t, indexes[1].Key, "name")
	assert.Contains(t, indexes[1].Key, "state.time")
	assert.Contains(t, indexes[2].Key, "correlationId")
	assert.Contains(t, indexes[3].Key, "state.time")
	assert.Equal(t, time.Duration(ttl)*time.Second, indexes[3].ExpireAfter)
}

func assertActionsAreNoLongerInDb(t *testing.T) {
	var got Action
	err := mongoT.FindOne(ActionCollectionId, bson.M{"_id": "1"}, &got)
	assert.Equal(t, "not found", err.Error())
	err = mongoT.FindOne(ActionCollectionId, bson.M{"_id": "2"}, &got)
	assert.Equal(t, "not found", err.Error())
	err = mongoT.FindOne(ActionCollectionId, bson.M{"_id": "3"}, &got)
	assert.Equal(t, "not found", err.Error())
}

func assertActionsAreStillInDb(t *testing.T) {
	var got Action
	mongoT.FindOne(ActionCollectionId, bson.M{"_id": "1"}, &got)
	assert.Equal(t, "1", got.Id)
	mongoT.FindOne(ActionCollectionId, bson.M{"_id": "2"}, &got)
	assert.Equal(t, "2", got.Id)
	mongoT.FindOne(ActionCollectionId, bson.M{"_id": "3"}, &got)
	assert.Equal(t, "3", got.Id)
}

func newActionT(id, name, state string, stateTime time.Time) Action {
	return Action{
		Id:    id,
		Name:  name,
		State: State{Value: state, Time: stateTime.Round(time.Millisecond)},
	}
}

type Action struct {
	Id    string `bson:"_id"`
	Name  string `bson:"name"`
	State State  `bson:"state"`
}

type State struct {
	Value string    `bson:"value"`
	Time  time.Time `bson:"time"`
}

func resetEnsureIndexes() {
	ensure = ensureIndexFn
}

func resetUpdateTTL() {
	update = updateTTLFn
}

func resetGetIndexes() {
	getIndexes = getIndexesFn
}
