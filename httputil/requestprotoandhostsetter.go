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
	"net/http"
)

// set the (possibly forwarded) http protocol and host in the request
func SetProtocolAndHostIn(r *http.Request) {
	r.Proto = getProtocol(r)
	r.Host = getHost(r)
}

func getProtocol(r *http.Request) string {
	protocol := r.Header.Get("X-Forwarded-Proto")
	if protocol == "" {
		protocol = getProtocolUsingTLS(r)
	}
	if protocol == "" {
		protocol = "http"
	}
	return protocol
}

func getProtocolUsingTLS(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	return ""
}

func getHost(r *http.Request) string {
	host := r.Header.Get("X-Flyte-Host")
	if host == "" {
		host = r.Header.Get("X-Forwarded-Host")
		if host == "" {
			host = r.Host
		}
	}
	return host
}
