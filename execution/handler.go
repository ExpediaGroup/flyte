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
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/husobee/vestigo"
	"github.com/rs/zerolog/log"
	"net/http"
)

func PostEvent(w http.ResponseWriter, r *http.Request) {

	packId := vestigo.Param(r, "packId")
	pack, err := packRepo.Get(packId)
	if err != nil {
		switch err {
		case PackNotFoundErr:
			log.Info().Msgf("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Err(err).Send()
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	pack.UpdateLastSeen()

	defer r.Body.Close()
	event, err := toEvent(*pack, r.Body)
	if err != nil {
		log.Err(err).Send()
		w.WriteHeader(http.StatusBadRequest)
		return
	}


	log.Info().Msgf("Received Event: EventName=%s Pack=%+v", event.Name, pack)
	log.Debug().Msgf("Event Contents: Event=%+v", event)

	flowSvc.HandleEvent(*event)
	w.WriteHeader(http.StatusAccepted)
}

func CompleteAction(w http.ResponseWriter, r *http.Request) {

	packId := vestigo.Param(r, "packId")
	pack, err := packRepo.Get(packId)
	if err != nil {
		switch err {
		case PackNotFoundErr:
			log.Info().Msgf("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Err(err).Send()
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	pack.UpdateLastSeen()

	defer r.Body.Close()
	result, err := toEvent(*pack, r.Body)
	if err != nil {
		log.Err(err).Send()
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info().Msgf("Received Event: EventName=%s Pack=%+v", result.Name, pack)
	log.Debug().Msgf("Event Contents: Event=%+v", result)

	actionId := vestigo.Param(r, "actionId")
	action, err := pack.CompleteAction(actionId, *result)
	if err != nil {
		switch err {
		case ActionNotFoundErr:
			log.Info().Msgf("Action actionId=%s packId=%s not found", actionId, pack.Id)
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Err(err).Msgf("Error completing actionId=%s with result=%+v", actionId, result)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	log.Info().
		Str("ActionId", action.Id).
		Str("CorrelationId", action.CorrelationId).
		Str("FlowName", action.FlowName).
		Str("PackName", action.PackName).
		Str("ActionName", action.Name).
		Str("StepId", action.StepId).
		Str("State", action.State.Value).
		Str("ResultEventPackId", action.Result.Pack.Id).
		Str("ResultEvent", action.Result.Name).
		Bool("ResultEventIsFatal", action.Result.isFatal()).
		Msg("Action completed")

	flowSvc.HandleEvent(*result)
	go flowSvc.HandleAction(*action)
	w.WriteHeader(http.StatusAccepted)
}

func TakeAction(w http.ResponseWriter, r *http.Request) {

	packId := vestigo.Param(r, "packId")
	pack, err := packRepo.Get(packId)
	if err != nil {
		switch err {
		case PackNotFoundErr:
			log.Info().Msgf("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Err(err).Send()
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	pack.UpdateLastSeen()

	actionName := r.FormValue("actionName")
	action, err := pack.TakeAction(actionName)

	if err != nil {
		log.Err(err).Msgf("Could not take action for packId=%s and actionName=%s", pack.Id, actionName)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if action == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	log.Info().Msgf("Action actionId=%s taken", action.Id)

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
	UpdateLastSeen(id string) error
}
