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
	"errors"
	"github.com/HotelsDotCom/flyte/collections"
	"github.com/HotelsDotCom/go-logger"
)

type Pack struct {
	Id     string            `bson:"_id"`
	Name   string            `bson:"name"`
	Labels map[string]string `bson:"labels,omitempty"`
}

func (p Pack) CompleteAction(actionId string, result Event) (*Action, error) {
	return completeAction(p, actionId, result)
}

var completeAction = completeActionFn

func completeActionFn(pack Pack, actionId string, result Event) (*Action, error) {

	action, err := actionRepo.Get(actionId)
	if err != nil || action == nil {
		return action, err
	}

	if action.PackName != pack.Name || !collections.ContainsAll(pack.Labels, action.PackLabels) {
		logger.Errorf("pack=%+v trying to complete actionId=%s which which it cannot handle", pack, action.Id)
		return nil, nil
	}
	return action, action.finish(result)
}

func (p Pack) TakeAction(actionName string) (*Action, error) {
	return takeAction(p, actionName)
}

var takeAction = takeActionFn

func takeActionFn(pack Pack, actionName string) (*Action, error) {

	action, err := actionRepo.FindNew(pack, actionName)
	if err != nil || action == nil {
		return action, err
	}

	return action, action.take()
}

func (p Pack) UpdateLastSeen() {
		updateLastSeen(p)
}

var updateLastSeen = updateLastSeenFn

func updateLastSeenFn(pack Pack) {
	err := packRepo.UpdateLastSeen(pack.Id)
	if err != nil {
		logger.Infof("error recording last seen record for a pack id %s: %v", pack.Id, err)
	}
}

var PackNotFoundErr = errors.New("pack not found")
