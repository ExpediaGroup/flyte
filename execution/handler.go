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
	"github.com/husobee/vestigo"
	"net/http"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/go-logger"
)

func PostEvent(w http.ResponseWriter, r *http.Request) {

	packId := vestigo.Param(r, "packId")
	pack, err := packRepo.Get(packId)
	if err != nil {
		switch err {
		case PackNotFoundErr:
			logger.Infof("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	defer r.Body.Close()
	event, err := toEvent(*pack, r.Body)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infof("Received Event: EventName=%s Pack=%+v", event.Name, pack)
	logger.Debugf("Event Contents: Event=%+v", event)

	go flowSvc.HandleEvent(*event)
	w.WriteHeader(http.StatusAccepted)
}

func CompleteAction(w http.ResponseWriter, r *http.Request) {

	packId := vestigo.Param(r, "packId")
	pack, err := packRepo.Get(packId)
	if err != nil {
		switch err {
		case PackNotFoundErr:
			logger.Infof("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	defer r.Body.Close()
	result, err := toEvent(*pack, r.Body)
	if err != nil {
		logger.Error(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	logger.Infof("Received Event: EventName=%s Pack=%+v", result.Name, pack)
	logger.Debugf("Event Contents: Event=%+v", result)

	actionId := vestigo.Param(r, "actionId")
	action, err := pack.CompleteAction(actionId, *result)
	if err != nil {
		switch err {
		case ActionNotFoundErr:
			logger.Infof("Action actionId=%s packId=%s not found", actionId, pack.Id)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Errorf("Error completing actionId=%s with result=%+v: %v", actionId, result, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logger.Infof("Action with actionId=%s has been completed, new state=%s", action.Id, action.State.Value)

	go flowSvc.HandleEvent(*result)
	go flowSvc.HandleAction(*action)
	w.WriteHeader(http.StatusAccepted)
}

func TakeAction(w http.ResponseWriter, r *http.Request) {

	packId := vestigo.Param(r, "packId")
	pack, err := packRepo.Get(packId)
	if err != nil {
		switch err {
		case PackNotFoundErr:
			logger.Infof("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	actionName := r.FormValue("actionName")
	action, err := pack.TakeAction(actionName)

	if err != nil {
		logger.Errorf("Could not take action for packId=%s and actionName=%s: %v", pack.Id, actionName, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if action == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	logger.Infof("Action actionId=%s taken", action.Id)

	httputil.WriteResponse(w, r, toActionResponse(r, packId, *action))
}

var flowSvc FlowService = flowService{}

type FlowService interface {
	HandleEvent(e Event)
	HandleAction(a Action)
}

var packRepo PackRepository = packMgoRepo{}

type PackRepository interface {
	Get(id string) (*Pack, error)
}
