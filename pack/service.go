package pack

import (
	"github.com/HotelsDotCom/go-logger"
	"github.com/jasonlvhit/gocron"
	"time"
)

/**
	time format should be "hh:mm" i.e. "23:00"
 */
func ScheduleDailyRemovalOfDeadPacksAt(time string, packGracePeriodInSeconds int) {
	s := gocron.NewScheduler()
	s.Every(1).Day().At(time).DoSafely(removePacksOlderThan, packGracePeriodInSeconds)
	<- s.Start()
}

func removePacksOlderThan(packGracePeriodInSeconds int) {
	date := getPastDateFrom(packGracePeriodInSeconds)

	info, err := packRepo.RemoveAllOlderThan(date)
	if err != nil {
		logger.Errorf("problem removing dead packs older than '%s'. err: '%v'.", date.String(), err)
		return
	}

	logger.Infof("%v dead packs older than '%s' removed.", info.Removed, date.String())
}

var getPastDateFrom = getPastDateFromFn
func getPastDateFromFn(secondsInPast int) time.Time {
	return currentDate().Add(time.Duration(-secondsInPast)*time.Second)
}

var currentDate = getCurrentDateFn
func getCurrentDateFn() time.Time {
	return time.Now()
}