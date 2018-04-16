// +build acceptance

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

package acceptancetest

import (
	"github.com/HotelsDotCom/flyte/acceptancetest/server"
	"github.com/HotelsDotCom/flyte/acceptancetest/urlutil"
	"github.com/HotelsDotCom/flyte/mongo/mongotest"
	"testing"
	"time"
)

type FeatureFile struct {
	name  string
	tests []Test
}

type Test struct {
	name     string
	testFunc func(t *testing.T)
}

var suite = []FeatureFile{
	{"datastore", DatastoreFeatures},
	{"execution", ExecutionFeatures},
	{"flow", FlowFeatures},
	{"info", InfoFeatures},
	{"pack", PackFeatures},
	{"auth", AuthFeatures},
	{"audit", AuditFeatures},
}

var mongoT *mongotest.MongoT
var flyteApi *server.Flyte
var dex *server.MockDex
var httpClient = urlutil.NewClient(15 * time.Second)

func TestFeatures(t *testing.T) {
	defer tearDown()
	startFlyte()

	for _, feature := range suite {
		t.Run(feature.name, func(t *testing.T) {
			for _, test := range feature.tests {
				t.Run(test.name, test.testFunc)
			}
		})
	}
}

func startFlyte() {
	mongoT = server.StartMongoT()
	dex = server.StartDex()
	flyteApi = server.StartFlyte(mongoT.GetUrl(), dex.IssuerURL())
}

func tearDown() {
	if flyteApi != nil {
		flyteApi.Stop()
	}
	if mongoT != nil {
		mongoT.Teardown()
	}
	if dex != nil {
		dex.Stop()
	}
}

func ResetFlyteApi(t *testing.T) {
	mongoT.DropDatabase(t)
}
