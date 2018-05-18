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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/husobee/vestigo"
)

type dataItemResponse struct {
	Key         string          `json:"key"`
	ContentType string          `json:"contentType"`
	Description string          `json:"description,omitempty"`
	Links       []httputil.Link `json:"links"`
}

type dataItemsResponse struct {
	Links     []httputil.Link    `json:"links"`
	DataItems []dataItemResponse `json:"datastore"`
}

func toDataItemsResponse(r *http.Request, dataItems []DataItem) dataItemsResponse {

	ds := []dataItemResponse{}
	for _, d := range dataItems {
		response := dataItemResponse{
			Key:         d.Key,
			ContentType: d.ContentType,
			Description: d.Description,
			Links:       []httputil.Link{{Href: httputil.UriBuilder(r).Path(flytepath.DatastorePath, d.Key).Build(), Rel: "self"}},
		}
		ds = append(ds, response)
	}

	defaultLinks := []httputil.Link{
		{Href: httputil.UriBuilder(r).Path(flytepath.DatastorePath).Build(), Rel: "self"},
		{Href: httputil.UriBuilder(r).Path(flytepath.DatastorePath).Parent().Build(), Rel: "up"},
		{Href: flytepath.GetUriDocPathFor(flytepath.DatastoreDoc), Rel: "help"},
	}
	return dataItemsResponse{
		DataItems: ds,
		Links:     defaultLinks,
	}
}

func toDataItem(r *http.Request) (DataItem, error) {

	fileContent, fileContentType, err := getMultipartFile(r)
	if err != nil {
		return DataItem{}, fmt.Errorf("error getting multipart file: %v", err)
	}

	key := vestigo.Param(r, "key")
	if key == "" {
		return DataItem{}, errors.New("data store item key is empty")
	}

	return DataItem{
		Key:         key,
		Description: r.Form.Get("description"),
		ContentType: fileContentType,
		Value:       fileContent,
	}, nil
}

func getMultipartFile(r *http.Request) (content []byte, contentType string, err error) {

	file, header, err := r.FormFile("value")
	if err != nil {
		err = fmt.Errorf("cannot parse multipart request: %v", err)
		return
	}
	defer file.Close()

	if content, err = ioutil.ReadAll(file); err != nil {
		err = fmt.Errorf("cannot read file content: %v", err)
		return
	} else if len(content) == 0 {
		err = errors.New("file content is empty")
		return
	}

	if contentType = header.Header.Get(httputil.HeaderContentType); contentType == "" {
		contentType = "text/plain; charset=us-ascii"
	}
	return
}
