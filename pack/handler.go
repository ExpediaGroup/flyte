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
	"github.com/husobee/vestigo"
	"net/http"
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/go-logger"
	"time"
)

func PostPack(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	pack := &Pack{}
	if err := json.NewDecoder(r.Body).Decode(pack); err != nil {
		logger.Errorf("Cannot convert request to pack: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	pack.generateId()
	pack.LastSeen = time.Now()

	if err := packRepo.Add(*pack); err != nil {
		logger.Errorf("Cannot save packName=%s, packLabels=%+v: %v", pack.Name, pack.Labels, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Infof("Pack registered: PackId=%s", pack.Id)
	w.Header().Set("Location", httputil.UriBuilder(r).Path(flytepath.PacksPath, pack.Id).Build())
	w.WriteHeader(http.StatusCreated)

	httputil.WriteResponse(w, r, toPackResponse(r, *pack))
}

func GetPacks(w http.ResponseWriter, r *http.Request) {

	packs, err := packRepo.FindAll()
	if err != nil {
		logger.Errorf("Cannot find packs: %v", err)
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
			logger.Infof("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Errorf("Cannot find packId=%s: %v", packId, err)
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
			logger.Infof("Pack packId=%s not found", packId)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Errorf("Cannot delete packId=%s: %v", packId, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logger.Infof("Pack PackId=%s deleted", packId)
	w.WriteHeader(http.StatusNoContent)
}
