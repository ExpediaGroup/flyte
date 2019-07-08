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

package execution

import (
	"github.com/HotelsDotCom/flyte/mongo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/mgo.v2/bson"
	"testing"
	"time"
)

func TestGet_ShouldReturnExistingPack(t *testing.T) {

	mongoT.DropDatabase(t)
	want := Pack{Id: "existingPack"}
	mongoT.Insert(t, mongo.PackCollectionId, want)

	got, err := packRepo.Get("existingPack")
	require.NoError(t, err)

	assert.Equal(t, want, *got)
}

func TestGet_ShouldReturnNilForNonExistingPack(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.PackCollectionId, Pack{Id: "existingPack"})

	_, err := packRepo.Get("nonExistingPack")

	assert.EqualError(t, err, PackNotFoundErr.Error())
}

func TestUpdateLastSeen_ShouldRecordLastSeenWithCurrentDate(t *testing.T) {

	mongoT.DropDatabase(t)
	mongoT.Insert(t, mongo.PackCollectionId, Pack{Id: "existingPack"})

	before := time.Now()
	err := packRepo.UpdateLastSeen("existingPack")
	require.NoError(t, err)

	var p map[string]interface{}
	err = mongoT.FindOne(mongo.PackCollectionId, bson.M{"_id": "existingPack"}, &p)
	require.NoError(t, err)
	require.NotNil(t, p)

	lastSeen := p["lastSeen"]
	require.NotNil(t, lastSeen)

	assert.WithinDuration(t, before, lastSeen.(time.Time), 1 * time.Second)
}
