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

package flytepath

import (
	"fmt"
	"net/http"
)

var uriMap = make(map[string]string)

func EnsureUriDocMapIsInitialised(r *http.Request) {
	if len(uriMap) == 0 {
		baseUri := fmt.Sprintf("%v://%v", r.Proto, r.Host)

		for key, value := range getFlyteDocPaths() {
			uriMap[key] = baseUri + value
		}
	}
}

func GetUriDocPathFor(name string) string {
	return uriMap[name]
}
