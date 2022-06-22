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
	"encoding/json"
	"fmt"
	"github.com/ExpediaGroup/flyte/flytepath"
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/husobee/vestigo"
	"github.com/rs/zerolog/log"
	"net/http"
	"regexp"
	"time"
)

var hateoasRegex, _ = regexp.Compile("up|self|/actionResult$|/takeAction$|/event$")

func PostPack(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	pack := &Pack{}

	if err := json.NewDecoder(r.Body).Decode(pack); err != nil {
		log.Err(err).Msg("Cannot convert request to pack")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := validateLinks(pack); err != nil {
		log.Err(err).Msg("invalid links found")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pack.generateId()
	pack.LastSeen = time.Now().UTC()

	if err := packRepo.Add(*pack); err != nil {
		log.Err(err).Msgf("Cannot save packName=%s, packLabels=%+v", pack.Name, pack.Labels)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Info().Msgf("Pack registered: PackId=%s", pack.Id)
	w.Header().Set("Location", httputil.UriBuilder(r).Path(flytepath.PacksPath, pack.Id).Build())
	w.WriteHeader(http.StatusCreated)

	httputil.WriteResponse(w, r, toPackResponse(r, *pack))
}

func GetPacks(w http.ResponseWriter, r *http.Request) {

	packs, err := packRepo.FindAll()
	if err != nil {
		log.Err(err).Msg("Cannot find packs")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputil.WriteResponse(w, r, toPacksResponse(r, packs))
}

func GetPack(w http.ResponseWriter, r *http.Request) {

	packId := vestigo.Param(r, "packId")
	pack, err := packRepo.Get(packId)

	if err != nil {
		switch err {
		case PackNotFoundErr:
			log.Info().Msgf("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Err(err).Msgf("Cannot find packId=%s", packId)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	httputil.WriteResponse(w, r, toPackResponse(r, *pack))
}

func DeletePack(w http.ResponseWriter, r *http.Request) {

	packId := vestigo.Param(r, "packId")

	if err := packRepo.Remove(packId); err != nil {
		switch err {
		case PackNotFoundErr:
			log.Info().Msgf("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			log.Err(err).Msgf("Cannot delete packId=%s", packId)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	log.Info().Msgf("Pack PackId=%s deleted", packId)
	w.WriteHeader(http.StatusNoContent)
}

func validateLinks(p *Pack) error {
	for _, link := range p.Links {
		if hateoasRegex.MatchString(link.Rel) {
			return fmt.Errorf("you can't use %s as it collides with flyte relative links", link.Rel)
		}
	}
	return nil
}
