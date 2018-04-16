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
	"math"
	"net/http"
	"strconv"
)

var (
	DefaultMaxLimit   = 300
	DefaultMinLimit   = 30
	DefaultPageNumber = 1
)

type Page struct {
	// defaults to 1, even if page does not contain any items
	PageNumber int
	// defaults to 30, max limit is 300
	PerPage int
	// it the page does not contain any items, this will be set to 0
	TotalPages int
	// (PageNumber - 1) * PerPage
	StartIndex int

	request *http.Request
}

// called in the handler, page and per_page are retrieved and set from
// the http request query
func NewPage(r *http.Request, totalItems int) *Page {

	pageNumber, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	if pageNumber < DefaultPageNumber {
		pageNumber = DefaultPageNumber
	}

	if perPage < 1 {
		perPage = DefaultMinLimit
	}

	if perPage > DefaultMaxLimit {
		perPage = DefaultMaxLimit
	}

	page := &Page{
		PageNumber: pageNumber,
		PerPage:    perPage,
		request:    r,
	}

	page.initialize(totalItems)
	return page
}

// called in the service, to initialize/populate 'TotalPages', 'StartIndex' ...
// Page members
func (p *Page) initialize(totalItems int) {

	if totalItems < 1 {
		p.TotalPages = 0
		p.PageNumber = 1
		p.StartIndex = 0
		return
	}

	// set number of available pages
	t := float64(totalItems) / float64(p.PerPage)
	p.TotalPages = int(math.Ceil(t))

	// validate page number
	if p.PageNumber > p.TotalPages {
		p.PageNumber = p.TotalPages
	}
	p.StartIndex = (p.PageNumber - 1) * p.PerPage
}

// return 'default' links, plus first, prev, next and last links (if applicable)
func (p *Page) PageLinksFor(uri string, defaultLinks []Link) []Link {

	pageTemplate := fmt.Sprintf("%s?page=%%d&per_page=%d", uri, p.PerPage)

	// first and prev
	if p.PageNumber != 1 && p.TotalPages > 1 {
		defaultLinks = append(defaultLinks, Link{Href: fmt.Sprintf(pageTemplate, 1), Rel: "first"})
		defaultLinks = append(defaultLinks, Link{Href: fmt.Sprintf(pageTemplate, p.PageNumber-1), Rel: "prev"})
	}

	// next and last
	if p.PageNumber != p.TotalPages && p.TotalPages > 1 {
		defaultLinks = append(defaultLinks, Link{Href: fmt.Sprintf(pageTemplate, p.PageNumber+1), Rel: "next"})
		defaultLinks = append(defaultLinks, Link{Href: fmt.Sprintf(pageTemplate, p.TotalPages), Rel: "last"})
	}

	return defaultLinks
}
