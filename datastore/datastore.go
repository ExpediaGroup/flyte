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

import "errors"

type DataItem struct {
	Key         string `bson:"_id"`
	ContentType string `bson:"contentType"`
	Description string `bson:"description"`
	Value       []byte `bson:"value"`
}

type Repository interface {
	Add(dataItem DataItem) error
	Remove(key string) error
	Get(key string) (*DataItem, error)
	FindAll() ([]DataItem, error)
	Has(key string) (bool, error)
}

var (
	dataItemNotFound = errors.New("not found")
	dataItemExists   = errors.New("data item already exists")
)
