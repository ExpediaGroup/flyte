// +build examples

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

package main

import (
	"bufio"
	"fmt"
	"github.com/HotelsDotCom/flyte-client/client"
	"github.com/HotelsDotCom/flyte-client/config"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/go-logger"
	"net/url"
	"os"
	"strings"
)

func main() {

	var conf = config.FromEnvironment()

	EventSuccessEventDef := flyte.EventDef{
		Name: "SendMessage",
	}

	packDef := flyte.PackDef{
		Name:      "MessageSender",
		HelpURL:   getUrl("https://github.com/HotelsDotCom/flyte/examples/README.md"),
		EventDefs: []flyte.EventDef{EventSuccessEventDef},
		Labels:    conf.Labels,
	}

	p := flyte.NewPack(packDef, client.NewClient(conf.FlyteApiUrl, conf.Timeout))

	p.Start()

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("-> ")
		text, _ := reader.ReadString('\n')
		// convert CRLF to LF
		text = strings.Replace(text, "\n", "", -1)

		if text == "quit" || text == "exit" {
			os.Exit(0)
		}
		p.SendEvent(flyte.Event{EventDef: EventSuccessEventDef, Payload: text})
	}
}

func getUrl(rawUrl string) *url.URL {
	url, err := url.Parse(rawUrl)
	if err != nil {
		logger.Fatalf("%s is not a valid url", rawUrl)
	}
	return url
}
