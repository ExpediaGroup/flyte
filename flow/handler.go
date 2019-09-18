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

package flow

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/go-logger"
	"github.com/husobee/vestigo"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

var flowRepo Repository = flowMgoRepo{}

const SchemaFile = "flow-schema.json"

func PostFlow(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	flow := Flow{}
	var bodyBytes []byte
	if r.Body != nil {
		bodyBytes, _ = ioutil.ReadAll(r.Body)
	}

	if err := validateJsonAgainstSchema(string(bodyBytes)); err != nil {
		logger.Errorf("Cannot convert request to flow: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := json.Unmarshal(bodyBytes, &flow); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		logger.Errorf("Cannot convert request to flow: %v", err)
		return
	}

	if err := flowRepo.Add(flow); err != nil {
		logger.Errorf("Cannot add flow to repo flowName=%s: %v", flow.Name, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Location", httputil.UriBuilder(r).Path(flytepath.FlowsPath, flow.Name).Build())
	w.WriteHeader(http.StatusCreated)
}

func GetFlows(w http.ResponseWriter, r *http.Request) {

	flows, err := flowRepo.FindAll()
	if err != nil {
		logger.Errorf("Cannot find flows: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	httputil.WriteResponse(w, r, toFlowsResponse(r, flows))
}

func GetFlow(w http.ResponseWriter, r *http.Request) {

	flowName := vestigo.Param(r, "flowName")
	flow, err := flowRepo.Get(flowName)
	if err != nil {
		switch err {
		case FlowNotFoundErr:
			logger.Infof("Flow flowName=%s not found", flowName)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Error("Cannot get flowName=%s: %v", flowName, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	httputil.WriteResponse(w, r, toFlowResponse(r, *flow))
}

func DeleteFlow(w http.ResponseWriter, r *http.Request) {

	flowName := vestigo.Param(r, "flowName")

	if err := flowRepo.Remove(flowName); err != nil {
		switch err {
		case FlowNotFoundErr:
			logger.Infof("Flow flowName=%s not found", flowName)
			w.WriteHeader(http.StatusNotFound)
		default:
			logger.Errorf("Cannot delete flowName=%s: %v", flowName, err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	logger.Infof("Flow flowName=%s deleted", flowName)
	w.WriteHeader(http.StatusNoContent)
}

var fileAs = filepath.Abs
var validate = gojsonschema.Validate

func validateJsonAgainstSchema(data string) error {
	schema, err := fileAs(SchemaFile)
	if err != nil {
		return err
	}
	loader := gojsonschema.NewReferenceLoader("file://" + schema)
	document := gojsonschema.NewStringLoader(data)
	result, err := validate(loader, document)
	if err != nil {
		switch err.(type) {
		case *os.PathError:
			return fmt.Errorf("file not found %s", loader.JsonSource())
		default:
			return err
		}
	}
	if result.Errors() != nil {
		return errors.New(result.Errors()[0].String())
	}
	return nil
}
