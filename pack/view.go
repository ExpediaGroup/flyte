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

package pack

import (
	"net/http"
	"github.com/HotelsDotCom/flyte/flytepath"
	"github.com/HotelsDotCom/flyte/httputil"
)

type packResponse struct {
	Pack
}

func toPackResponse(r *http.Request, pack Pack) packResponse {

	pr := packResponse{pack}
	for i := range pr.Commands {
		link := httputil.Link{Href: httputil.UriBuilder(r).
			Path(flytepath.TakeActionWithCommandPath).
			Replace(":packId", pack.Id).
			Replace(":commandName", pack.Commands[i].Name).
			Build(),
			Rel: flytepath.GetUriDocPathFor(flytepath.TakeActionDoc)}

		pr.Commands[i].Links = []httputil.Link{link}
	}

	pr.SetStatus()

	pr.Links = append(pr.Links, httputil.Link{Href: httputil.UriBuilder(r).Path(flytepath.PackPath).Replace(":packId", pack.Id).Build(), Rel: "self"})
	pr.Links = append(pr.Links, httputil.Link{Href: httputil.UriBuilder(r).Path(flytepath.PackPath).Parent().Build(), Rel: "up"})
	pr.Links = append(pr.Links, httputil.Link{Href: httputil.UriBuilder(r).Path(flytepath.TakeActionPath).Replace(":packId", pack.Id).Build(), Rel: flytepath.GetUriDocPathFor(flytepath.TakeActionDoc)})
	pr.Links = append(pr.Links, httputil.Link{Href: httputil.UriBuilder(r).Path(flytepath.PostEventPath).Replace(":packId", pack.Id).Build(), Rel: flytepath.GetUriDocPathFor(flytepath.PostEventDoc)})
	return pr
}

type packsResponse struct {
	Links []httputil.Link `json:"links"`
	Packs []packResponse  `json:"packs"`
}

func toPacksResponse(r *http.Request, packs []Pack) packsResponse {

	ps := []packResponse{}
	for _, p := range packs {
		pr := packResponse{p}
		pr.Links = []httputil.Link{{Href: httputil.UriBuilder(r).Path(flytepath.PackPath).Replace(":packId", p.Id).Build(), Rel: "self"}}
		pr.SetStatus()
		ps = append(ps, pr)
	}

	defaultLinks := []httputil.Link{
		{Href: httputil.UriBuilder(r).Path(flytepath.PacksPath).Build(), Rel: "self"},
		{Href: httputil.UriBuilder(r).Path(flytepath.PacksPath).Parent().Build(), Rel: "up"},
		{Href: flytepath.GetUriDocPathFor(flytepath.GetPacksDoc), Rel: "help"},
	}
	return packsResponse{
		Packs: ps,
		Links: defaultLinks,
	}
}
