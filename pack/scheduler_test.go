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
	"fmt"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

const oneWeekInSeconds = 604800

func TestRemovePacksOlderThan_ShouldPassInTheExpectedDateToTheRepoFunction(t *testing.T) {

	// given our 'getPastDateFrom' function will return 'oneWeekAgo' date
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	defer resetGetDateFrom()
	getPastDateFrom = func(secondsInPast int) time.Time {
		return oneWeekAgo
	}

	// and we are recording the date passed in to our (mock) repo
	var passedInDate time.Time
	defer resetPackRepo()
	packRepo = mockPackRepo{
		removeAllOlderThan: func(date time.Time) (packsRemoved int, err error) {
			passedInDate = date
			return 2, nil
		},
	}

	// when
	removePacksOlderThan(oneWeekInSeconds)

	// then
	assert.Equal(t, oneWeekAgo, passedInDate)
}

func TestRemovePacksOlderThan_ShouldLogOnError(t *testing.T) {

	// given
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	// and the (mock) repo is returning an error
	var passedInDate time.Time
	defer resetPackRepo()
	packRepo = mockPackRepo{
		removeAllOlderThan: func(date time.Time) (packsRemoved int, err error) {
			passedInDate = date
			return 0, errors.New("some error")
		},
	}

	// when
	removePacksOlderThan(oneWeekInSeconds)

	// then
	expectedMessage := fmt.Sprintf("problem removing dead packs older than '%s'. err: 'some error'.", passedInDate.Format(time.RFC850))
	assert.Equal(t, expectedMessage, loggertest.GetLogMessages()[0].Message)
}

func TestGetPastDateFrom_ShouldReturnExpectedPastDateFromTheValueInSecondsPassedIn(t *testing.T) {

	// given the 'currentDate' function will return the 'now' date we have created here
	now := time.Now()
	defer resetGetCurrentDate()
	currentDate = func() time.Time {
		return now
	}

	// when we pass in a time period of one week in seconds
	pastDate := getPastDateFrom(oneWeekInSeconds)

	// then the date we expect is exactly one week in the past (using the date we created above)
	expectedDate := now.AddDate(0, 0, -7)
	assert.True(t, expectedDate.Equal(pastDate))
}

func resetGetDateFrom() {
	getPastDateFrom = getPastDateFromFn
}

func resetGetCurrentDate() {
	currentDate = getCurrentDateFn
}
