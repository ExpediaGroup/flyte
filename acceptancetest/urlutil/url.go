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

package urlutil

import (
	"fmt"
	"github.com/HotelsDotCom/flyte/httputil"
	"strings"
)

func FindURLByRel(links []httputil.Link, rel string) (string, error) {
	for _, l := range links {
		if strings.HasSuffix(l.Rel, rel) {
			return l.Href, nil
		}
	}
	return "", fmt.Errorf("Could not find link with rel %q in %v", rel, links)
}

func Join(path string, elements ...string) string {

	if len(elements) == 0 {
		return path
	}

	path = strings.TrimSuffix(path, "/")
	for i, e := range elements {
		if e == "" {
			continue
		}
		e = strings.TrimPrefix(e, "/")
		if i != len(elements)-1 {
			e = strings.TrimSuffix(e, "/")
		}
		path = path + "/" + e
	}
	return path
}
