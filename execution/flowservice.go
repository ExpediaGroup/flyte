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
	"github.com/HotelsDotCom/go-logger"
)

type flowService struct{}

func (flowService) HandleEvent(e Event) {

	flows, err := flowRepo.FindByEvent(e)
	if err != nil {
		logger.Errorf("Error handling event=%+v: %v", e, err)
		return
	}

	for _, f := range flows {
		go func(f Flow) {
			f.HandleEvent(e)
		}(f)
	}
}

func (flowService) HandleAction(a Action) {

	flow, err := flowRepo.GetByAction(a)
	if err != nil {
		logger.Errorf("Error handling action=%+v: %v", a, err)
		return
	} else if flow == nil {
		logger.Errorf("Error handling action=%+v: flow not found", a)
		return
	}

	flow.HandleEvent(a.Result)
}

var flowRepo FlowRepository = flowMgoRepo{}

type FlowRepository interface {
	GetByAction(a Action) (*Flow, error)
	FindByEvent(e Event) ([]Flow, error)
}
