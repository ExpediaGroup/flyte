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
	"os"
	"github.com/HotelsDotCom/flyte/mongo"
	"github.com/HotelsDotCom/flyte/mongo/mongotest"
	"testing"
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
