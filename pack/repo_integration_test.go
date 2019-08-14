// +build integration

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2/bson"
	"os"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/flyte/mongo"
	"github.com/HotelsDotCom/flyte/mongo/mongotest"
	"testing"
	"time"
)

const ttl = 365 * 24 * 60 * 60

var mongoT *mongotest.MongoT

func TestMain(m *testing.M) {
	os.Exit(runTestsWithMongo(m))
}

func runTestsWithMongo(m *testing.M) int {
	mongoT = mongotest.NewMongoT(mongo.DbName)
	defer mongoT.Teardown()

	mongoT.Start()

	mongo.InitSession(mongoT.GetUrl(), ttl)

	return m.Run()
}

var slackPack = Pack{
	Id:     "Slack.env.prod",
	Name:   "Slack",
	Labels: map[string]string{"env": "prod"},
	Commands: []Command{
		{
			Name:   "SendMessage",
			Events: []string{"MessageSent", "SendMessageFailed"},
			Links:  []httputil.Link{{Href: "http://flyte.pack/slack/commands/help", Rel: "help"}},
		},
	},
	LastSeen: time.Now(),
	Events: []Event{
		{Name: "MessageSent", Links: []httputil.Link{{Href: "http://flyte.pack/slack/events/help", Rel: "help"}}},
		{Name: "SendMessageFailed", Links: []httputil.Link{{Href: "http://flyte.pack/slack/events/help", Rel: "help"}}},
	},
	Links: []httputil.Link{
		{Href: "http://flyte.pack/slack/pack/help", Rel: "help"},
	},
}

func TestAdd_ShouldAddPackIntoRepo(t *testing.T) {
	mongoT.DropDatabase(t)

	err := packRepo.Add(slackPack)
	require.NoError(t, err)

	var p Pack
	mongoT.FindOneT(t, mongo.PackCollectionId, bson.M{"_id": slackPack.Id}, &p)
	assert.Equal(t, slackPack.Name, p.Name)
	assert.Equal(t, slackPack.Labels, p.Labels)
	assert.Equal(t, slackPack.Commands, p.Commands)
	assert.Equal(t, slackPack.Events, p.Events)
	assert.Equal(t, slackPack.Links, p.Links)
	assert.WithinDuration(t, slackPack.LastSeen, p.LastSeen, 1 * time.Second)
}

func TestRemove_ShouldRemovePackFromRepo(t *testing.T) {
	mongoT.DropDatabase(t)
	insertPack(t, slackPack)

	err := packRepo.Remove(slackPack.Id)
	require.NoError(t, err)

	packs, _ := packRepo.FindAll()
	assert.Equal(t, len(packs), 0)
}

func TestGet_ShouldGetPackFromRepo(t *testing.T) {
	mongoT.DropDatabase(t)
	insertPack(t, slackPack)

	p, err := packRepo.Get(slackPack.Id)
	require.NoError(t, err)

	require.NotNil(t, p, "pack returned should not be nil")
	assert.Equal(t, slackPack.Name, p.Name)
	assert.Equal(t, slackPack.Labels, p.Labels)
	assert.Equal(t, slackPack.Commands, p.Commands)
	assert.Equal(t, slackPack.Events, p.Events)
	assert.Equal(t, slackPack.Links, p.Links)
	assert.WithinDuration(t, slackPack.LastSeen, p.LastSeen, 1 * time.Second)
}

func TestFindAll_ShouldReturnAllPacksFromRepo(t *testing.T) {
	mongoT.DropDatabase(t)
	insertPack(t, slackPack)
	hipChatPack := Pack{Name: "Hipchat", Labels: map[string]string{"env": "dev"}, LastSeen:time.Now()}
	insertPack(t, hipChatPack)

	packs, err := packRepo.FindAll()
	require.NoError(t, err)

	assert.Equal(t, 2, len(packs))
	assert.Equal(t, slackPack.Name, packs[1].Name)
	assert.Equal(t, slackPack.Labels, packs[1].Labels)
	assert.WithinDuration(t, slackPack.LastSeen, packs[1].LastSeen, 1 * time.Second)

	assert.Equal(t, hipChatPack.Name, packs[0].Name)
	assert.Equal(t, hipChatPack.Labels, packs[0].Labels)
	assert.WithinDuration(t, hipChatPack.LastSeen, packs[0].LastSeen, 1 * time.Second)
}

func TestDeleteAllOlderThan_ShouldRemovePacksFromRepoOlderThanTheDatePassedIn(t *testing.T) {
	now := time.Now()
	oneWeekAgo := now.AddDate(0, 0, -7)
	eightDaysAgo := now.AddDate(0, 0, -8)
	oneMonthAgo := now.AddDate(0, -1, 0)

	mongoT.DropDatabase(t)
	insertPack(t, Pack{
		Id:     "Slack.env.prod",
		Name:   "Slack",
		LastSeen: now,
	})
	insertPack(t, Pack{
		Id:     "Argo.env.prod",
		Name:   "Argo",
		LastSeen: eightDaysAgo,
	})
	insertPack(t, Pack{
		Id:     "Bamboo.env.prod",
		Name:   "Bamboo",
		LastSeen: oneMonthAgo,
	})

	info, err := packRepo.RemoveAllOlderThan(oneWeekAgo)
	require.NoError(t, err)

	assert.Equal(t, 2, info.Removed)
	packs, _ := packRepo.FindAll()
	assert.Equal(t, 1, len(packs))
	assert.Equal(t, "Slack", packs[0].Id)
}

func insertPack(t *testing.T, pack Pack) {
	pack.generateId()
	mongoT.UpsertId(t, mongo.PackCollectionId, pack.Id, pack)
}
