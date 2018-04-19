package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"github.com/HotelsDotCom/flyte-client/client"
	"github.com/HotelsDotCom/flyte-client/config"
	"github.com/HotelsDotCom/flyte-client/flyte"
	"github.com/HotelsDotCom/go-logger"
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
