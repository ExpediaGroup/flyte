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

package collections

import (
	"fmt"
	"sort"
)

func ContainsAll(superSet, subSet map[string]string) bool {
	for k, v := range subSet {
		if v2, ok := superSet[k]; !ok || v2 != v {
			return false
		}
	}
	return true
}

func Merge(maps ...map[string]string) map[string]string {
	merged := map[string]string{}
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

func SortedKeys(m map[string]string) []string {

	keys := []string{}
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func Contains(slice []string, contain string) bool {
	for _, v := range slice {
		if v == contain {
			return true
		}
	}
	return false
}

// checks if the slices have at least one matching element i.e. their intersection is non-nil
func HasMatchingElement(a []string, b []string) bool {
	for _, aVal := range a {
		for _, bVal := range b {
			if aVal == bVal {
				return true
			}
		}
	}
	return false
}

func ToStringSlice(a []interface{}) ([]string, error) {
	converted := make([]string, len(a))
	for i, v := range a {
		switch val := v.(type) {
		case string:
			converted[i] = val
		default:
			return nil, fmt.Errorf("can only convert slices of type 'string' - slices of type %T are not supported", val)
		}
	}
	return converted, nil
}
