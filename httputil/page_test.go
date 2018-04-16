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
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewPageDefault(t *testing.T) {

	cases := []struct {
		page               *Page
		expectedPageNumber int
		expectedPerPage    int
	}{
		{NewPage(getRequest("", ""), 100), 1, 30},
		{NewPage(getRequest("-3", "1000"), 100), 1, 300},
		{NewPage(getRequest("x", "y"), 100), 1, 30},
	}

	for _, c := range cases {
		if c.expectedPageNumber != c.page.PageNumber {
			t.Errorf("\npage number\nExpected: %d\nActual:   %d", c.expectedPageNumber, c.page.PageNumber)
		}
		if c.expectedPerPage != c.page.PerPage {
			t.Errorf("\nper page\nExpected: %d\nActual:   %d", c.expectedPerPage, c.page.PerPage)
		}
	}
}

func TestPage_Initialize(t *testing.T) {

	cases := []struct {
		page               *Page
		expectedTotalPages int
		expectedPageNumber int
		expectedStartIndex int
	}{
		{NewPage(getRequest("1", "5"), 10), 2, 1, 0},
		{NewPage(getRequest("2", "5"), 0), 0, 1, 0},
	}

	for _, c := range cases {
		if c.expectedTotalPages != c.page.TotalPages {
			t.Errorf("\ntotal pages\nExpected: %d\nActual:   %d", c.expectedTotalPages, c.page.TotalPages)
		}
		if c.expectedPageNumber != c.page.PageNumber {
			t.Errorf("\npage number\nExpected: %d\nActual:   %d", c.expectedPageNumber, c.page.PageNumber)
		}
		if c.expectedStartIndex != c.page.StartIndex {
			t.Errorf("\nstart index\nExpected: %d\nActual:   %d", c.expectedStartIndex, c.page.StartIndex)
		}
	}
}

func TestPageLinks(t *testing.T) {

	request := getRequest("", "")
	SetProtocolAndHostIn(request)
	page := NewPage(request, 100)
	packsUri := UriBuilder(request).Path("packs").Build()
	defaultLinks := []Link{
		{Href: packsUri, Rel: "self"},
		{Href: UriBuilder(request).Path("").Build(), Rel: "up"},
		{Href: UriBuilder(request).Path("swagger#/packs").Build(), Rel: "help"},
	}
	links := page.PageLinksFor(packsUri, defaultLinks)

	assert.Equal(t, "http://flyte/packs", links[0].Href)
	assert.Equal(t, "self", links[0].Rel)
	assert.Equal(t, "http://flyte/", links[1].Href)
	assert.Equal(t, "up", links[1].Rel)
	assert.Equal(t, "http://flyte/swagger#/packs", links[2].Href)
	assert.Equal(t, "help", links[2].Rel)
	assert.Equal(t, "http://flyte/packs?page=2&per_page=30", links[3].Href)
	assert.Equal(t, "next", links[3].Rel)
	assert.Equal(t, "http://flyte/packs?page=4&per_page=30", links[4].Href)
	assert.Equal(t, "last", links[4].Rel)
}

func getRequest(page, perPage string) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "http://flyte", nil)

	query := []string{}
	if page != "" {
		query = append(query, "page="+page)
	}
	if perPage != "" {
		query = append(query, "per_page="+perPage)
	}

	r.URL.RawQuery = strings.Join(query, "&")
	return r
}
