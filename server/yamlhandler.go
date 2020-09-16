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

package server

import (
	"bytes"
	"github.com/ExpediaGroup/flyte/httputil"
	"github.com/HotelsDotCom/go-logger"
	"github.com/ghodss/yaml"
	"io/ioutil"
	"net/http"
)

func YamlHandler(h http.HandlerFunc) http.HandlerFunc {
	return yamlHandler(h)
}

var yamlHandler = func(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := yamlToJSONHandler(w, r); err != nil {
			return
		}
		h(w, r)
	}
}

func yamlToJSONHandler(w http.ResponseWriter, r *http.Request) error {
	if isYAMLContentType(r) {
		if err := convertYAMLRequestToJSONRequest(r); err != nil {
			logger.Errorf("cannot process yaml request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return err
		}
	}
	return nil
}

func isYAMLContentType(r *http.Request) bool {
	return r.Header.Get(httputil.HeaderContentType) == httputil.MediaTypeYaml
}

func convertYAMLRequestToJSONRequest(r *http.Request) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	body, err := yaml.YAMLToJSON(data)
	if err != nil {
		return err
	}

	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	r.Header.Set(httputil.HeaderContentType, httputil.MediaTypeJson)

	return nil
}
