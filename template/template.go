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

package template

import (
	"fmt"
	"reflect"
)

type Context map[string]interface{}

// static context will be added to the context passed to Resolve on each request
var staticContext = Context{}

func AddStaticContextEntry(key string, value interface{}) {
	staticContext[key] = value
}

// Creates a deep copy of whatever is passed to it
// and evaluates string nodes as templates with the provided context.
// This function is intended to be used with JSON objects, so does not handle pointers and structs.
func Resolve(v interface{}, c Context) (out interface{}, err error) {

	if v == nil {
		return v, nil
	}
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("error while evaluating expression: '%+v': %v", v, r)
		}
	}()

	out = resolveValue(reflect.ValueOf(v), c).Interface()

	return out, nil
}

func resolveValue(v reflect.Value, context Context) reflect.Value {

	switch v.Kind() {

	case reflect.Ptr, reflect.Struct:
		panic(fmt.Errorf("unsupported kind: %v", v.Kind()))

	case reflect.Interface:
		return resolveValue(v.Elem(), context)

	case reflect.Slice:
		s := reflect.MakeSlice(v.Type(), v.Len(), v.Cap())
		for i := 0; i < s.Len(); i++ {
			s.Index(i).Set(resolveValue(v.Index(i), context))
		}
		return s

	case reflect.Map:
		m := reflect.MakeMap(v.Type())
		for _, k := range v.MapKeys() {
			m.SetMapIndex(k, resolveValue(v.MapIndex(k), context))
		}
		return m

	case reflect.String:
		s, err := execute(v.Interface().(string), context)
		if err != nil {
			panic(err)
		}
		return reflect.ValueOf(s)
	}

	return v
}
