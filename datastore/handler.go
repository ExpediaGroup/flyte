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

package datastore

import (
	"encoding/json"
	"fmt"
	"github.com/husobee/vestigo"
	"net/http"
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/go-logger"
	"strings"
)

var datastoreRepo Repository = datastoreMgoRepo{}

func GetItems(w http.ResponseWriter, r *http.Request) {

	dataItems, err := datastoreRepo.FindAll()
	if err != nil {
		logger.Errorf("cannot retrieve data items: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputil.WriteResponse(w, r, toDataItemsResponse(r, dataItems))
}

func GetItem(w http.ResponseWriter, r *http.Request) {

	key := vestigo.Param(r, "key")
	dataItem, err := datastoreRepo.Get(key)
	if err != nil {
		switch err {
		case dataItemNotFound:
			logger.Errorf("Data item key=%s not found", key)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Errorf("Cannot retrieve data item key=%s: %v", key, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set(httputil.HeaderContentType, dataItem.ContentType)
	w.Write(dataItem.Value)
}

func PutItem(w http.ResponseWriter, r *http.Request) {
	item, err := toDataItem(r)
	if err != nil {
		logger.Errorf("Error putting data store item: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isUpdate, err := datastoreRepo.Has(item.Key)
	if err != nil {
		logger.Errorf("Cannot check if item exists key=%s: %v", item.Key, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := datastoreRepo.Add(item); err != nil {
		logger.Errorf("Cannot add item key=%s: %v", item.Key, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	logger.Infof("Data item stored: key=%s contentType=%s", item.Key, item.ContentType)
	w.Header().Set("Location", httputil.UriBuilder(r).Path(flytepath.DatastorePath, item.Key).Build())

	if isUpdate {
		w.WriteHeader(http.StatusNoContent)
	} else {
		w.WriteHeader(http.StatusCreated)
	}
}

func DeleteItem(w http.ResponseWriter, r *http.Request) {

	key := vestigo.Param(r, "key")
	if err := datastoreRepo.Remove(key); err != nil {
		switch err {
		case dataItemNotFound:
			logger.Errorf("Data item key=%s not found", key)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Errorf("Cannot delete item key=%s: %v", key, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logger.Infof("Deleted data item key=%s", key)
	w.WriteHeader(http.StatusNoContent)
}

func GetDataStoreValue(key string) (interface{}, error) {

	dataItem, err := datastoreRepo.Get(key)
	if err != nil {
		return nil, fmt.Errorf("cannot find datastore item key=%s: %v", key, err)
	}

	if !strings.HasPrefix(dataItem.ContentType, httputil.MediaTypeJson) &&
		!strings.HasPrefix(dataItem.ContentType, "text/json") {
		return string(dataItem.Value), nil
	}

	value := map[string]interface{}{}
	if err := json.Unmarshal(dataItem.Value, &value); err != nil {
		return nil, fmt.Errorf("cannot unmarshal datastore item key=%s: %v", key, err)
	}
	return value, nil
}
