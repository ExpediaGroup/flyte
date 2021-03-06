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

package pack

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestScheduleDailyRemovalOfDeadPacksAt_ShouldCallFunctionToRemovePacksAtExpectedTime(t *testing.T) {
	// given we have a mock repo that will return a successful result if called
	repoCalled := false
	defer resetPackRepo()
	packRepo = mockPackRepo{
		removeAllOlderThan: func(date time.Time) (packsRemoved int, err error) {
			repoCalled = true
			return 2, nil
		},
	}

	// and we want to schedule a removal of packs one minute from now
	oneMinFromNow := time.Now().Add(time.Minute * time.Duration(1))
	oneMinFromNowFormatted := oneMinFromNow.Format("15:04")

	// when we call the scheduler and give it time to run its job
	s, sc := ScheduleDailyRemovalOfDeadPacksAt(oneMinFromNowFormatted, 1000)
	time.Sleep(time.Duration(61) * time.Second)

	// then we simply assert that the repo was called
	assert.True(t, repoCalled)
	// and clean up
	s.Clear()
	close(sc)
}
