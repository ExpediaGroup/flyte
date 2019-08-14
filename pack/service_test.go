package pack

import (
	"fmt"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"testing"
	"time"
)

const oneWeekInSeconds = 604800

func TestScheduleDailyRemovalOfDeadPacksAt_ShouldCallFunctionToRemovePacksAtExpectedTime(t *testing.T) {
	//oneMinFromNow := time.Now().Add(time.Minute * time.Duration(1))
	//
	//// record date passed in...
	//var passedInDate time.Time
	//defer resetPackRepo()
	//packRepo = mockPackRepo{
	//	removeAllOlderThan: func(date time.Time) (info *mgo.ChangeInfo, err error) {
	//		passedInDate = date
	//		return &mgo.ChangeInfo{Removed:2}, nil
	//	},
	//}



}

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
		removeAllOlderThan: func(date time.Time) (info *mgo.ChangeInfo, err error) {
			passedInDate = date
			return &mgo.ChangeInfo{Removed:2}, nil
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
		removeAllOlderThan: func(date time.Time) (info *mgo.ChangeInfo, err error) {
			passedInDate = date
			return nil, errors.New("some error")
		},
	}

	// when
	removePacksOlderThan(oneWeekInSeconds)

	// then
	expectedMessage := fmt.Sprintf("problem removing dead packs older than '%s'. err: 'some error'.", passedInDate.String())
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