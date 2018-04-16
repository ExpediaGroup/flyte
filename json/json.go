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

package json

import (
	"encoding/json"
	"fmt"
	"io"
)

// deserialize
func NewJson(r io.Reader) (Json, error) {
	var value Json
	d := json.NewDecoder(r)
	d.UseNumber()
	err := d.Decode(&value)
	if err != nil {
		err = fmt.Errorf("could not create Json from reader: %s", err)
		return nil, err
	}
	return value, nil
}

// Underlying value will be of type:
//
// - bool, for JSON booleans
// - float64, for JSON floats
// - Number, for JSON numbers
// - string, for JSON string literals
// - []interface{} for JSON arrays
// - map[string]interface{} for JSON objects
// - nil, for JSON null
//
type Json interface{}
