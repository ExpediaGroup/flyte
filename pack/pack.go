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
	"errors"
	"fmt"
	"github.com/HotelsDotCom/flyte/collections"
	"github.com/HotelsDotCom/flyte/httputil"
)

type Pack struct {
	Id       string            `json:"id" bson:"_id"`
	Name     string            `json:"name"`
	Labels   map[string]string `json:"labels,omitempty"`
	Commands []Command         `json:"commands,omitempty"`
	Events   []Event           `json:"events,omitempty"`
	Links    []httputil.Link   `json:"links,omitempty"`
}

type Command struct {
	Name   string          `json:"name"`
	Events []string        `json:"events"`
	Links  []httputil.Link `json:"links,omitempty"`
}

type Event struct {
	Name  string          `json:"name"`
	Links []httputil.Link `json:"links,omitempty"`
}

func (p *Pack) generateId() {

	id := p.Name
	for _, k := range collections.SortedKeys(p.Labels) {
		id += fmt.Sprintf(".%s.%s", k, p.Labels[k])
	}
	p.Id = id
}

type Repository interface {
	Add(pack Pack) error
	Remove(id string) error
	Get(id string) (*Pack, error)
	FindAll() ([]Pack, error)
}

var PackNotFoundErr = errors.New("pack not found")
