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
	"github.com/HotelsDotCom/go-logger"
	"github.com/jasonlvhit/gocron"
	"time"
)

/**
	time format should be "HH:MM" i.e. "23:00"
 */
func ScheduleDailyRemovalOfDeadPacksAt(time string, packGracePeriodInSeconds int) (*gocron.Scheduler, chan bool) {
	s := gocron.NewScheduler()
	s.Every(1).Day().At(time).Do(removePacksOlderThan, packGracePeriodInSeconds)
	sc := s.Start()
	return s, sc
}

func removePacksOlderThan(packGracePeriodInSeconds int) {
	date := getPastDateFrom(packGracePeriodInSeconds)

	packsRemoved, err := packRepo.RemoveAllOlderThan(date)
	if err != nil {
		logger.Errorf("problem removing dead packs older than '%s'. err: '%v'.", date.Format(time.RFC850), err)
		return
	}

	if packsRemoved > 0 {
		logger.Infof("%v dead pack/s older than '%s' removed.", packsRemoved, date.Format(time.RFC850))
	}
}

var getPastDateFrom = getPastDateFromFn
func getPastDateFromFn(secondsInPast int) time.Time {
	return currentDate().Add(time.Duration(-secondsInPast)*time.Second)
}

var currentDate = getCurrentDateFn
func getCurrentDateFn() time.Time {
	return time.Now()
}