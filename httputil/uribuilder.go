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

package httputil

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
)

type Builder struct {
	baseUri string
	path    string
}

// works out the request uri, including scheme, and port from the request passed in
// the full path is built up from this base
func UriBuilder(r *http.Request) *Builder {
	baseUri := fmt.Sprintf("%v://%v/", r.Proto, r.Host)
	return &Builder{baseUri: baseUri}
}

// the path/s to be added to the base uri, in the order passed in
func (b *Builder) Path(path ...string) *Builder {
	b.path = filepath.Join(path...)
	return b
}

func (b *Builder) Parent() *Builder {
	p := filepath.Dir(b.path)
	if p == "." {
		p = "/"
	}
	b.path = p
	return b
}

// replaces a path parameter in the currently built path with the value passed in. will also remove a 'slash' suffix if one exists.
func (b *Builder) Replace(pathParam, paramValue string) *Builder {
	p := strings.Replace(b.path, pathParam, paramValue, 1)
	b.path = strings.TrimSuffix(p, "/")
	return b
}

func (b *Builder) Build() string {
	return b.baseUri + strings.TrimPrefix(b.path, "/")
}
