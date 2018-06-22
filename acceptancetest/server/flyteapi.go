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

package server

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"github.com/HotelsDotCom/flyte/mongo"
	"github.com/HotelsDotCom/flyte/mongo/mongotest"
	"github.com/HotelsDotCom/flyte/acceptancetest/urlutil"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/flyte/server"
	"github.com/HotelsDotCom/go-logger"
	"time"
)

const ttl = 365 * 24 * 60 * 60

// use StartFlyte to create an instance which sets all URLs
type Flyte struct {
	rootURL           *url.URL
	flowsURL          string
	packsURL          string
	healthURL         string
	datastoreURL      string
	flowExecutionsURL string
	swaggerURL        string

	certFilePath string
	keyFilePath  string
	port         string

	server *server.FlyteServer
}

func StartFlyte(mgoHost, oidcIssuerUri string) *Flyte {
	flyte, err := startFlyte(mgoHost, oidcIssuerUri)
	if err != nil {
		logger.Fatalf("Unable to start flyte: %v", err)
	}
	return flyte
}

func startFlyte(mgoHost, oidcIssuerUri string) (*Flyte, error) {
	flyteapi := &Flyte{}
	if err := flyteapi.setPort(); err != nil {
		return flyteapi, err
	}

	if err := flyteapi.writeCertAndKeyFile(); err != nil {
		flyteapi.deleteCertAndKeyFile()
		return flyteapi, err
	}

	links, err := flyteapi.startFlyteApi(mgoHost, oidcIssuerUri)
	if err != nil {
		return flyteapi, err
	}

	if err := flyteapi.discoverURLs(links); err != nil {
		return flyteapi, err
	}
	return flyteapi, nil
}

func StartMongoT() *mongotest.MongoT {
	mongoT := mongotest.NewMongoT(mongo.DbName)
	mongoT.Start()
	return mongoT
}

func (f *Flyte) writeCertAndKeyFile() error {

	f.certFilePath = fmt.Sprintf("%scert.pem", os.TempDir())
	if err := ioutil.WriteFile(f.certFilePath, []byte(flyteapiCert), 0644); err != nil {
		return err
	}

	f.keyFilePath = fmt.Sprintf("%skey.pem", os.TempDir())
	if err := ioutil.WriteFile(f.keyFilePath, []byte(flyteapiKey), 0644); err != nil {
		return err
	}

	return nil
}

func (f *Flyte) deleteCertAndKeyFile() {

	if err := os.Remove(f.certFilePath); err != nil && f.certFilePath != "" {
		panic(err)
	}
	if err := os.Remove(f.keyFilePath); err != nil && f.keyFilePath != "" {
		panic(err)
	}
}

func (f *Flyte) startFlyteApi(mgoHost, oidcIssuerUri string) (map[string][]httputil.Link, error) {

	f.server = server.NewFlyteServer(f.port, mgoHost, ttl)
	f.server.EnableAuth("./testdata/policy_config.yaml", oidcIssuerUri, "example-app")
	go f.server.ListenAndServeTLS(f.certFilePath, f.keyFilePath)
	time.Sleep(500 * time.Millisecond) // wait a bit for server to start

	baseURL := getBaseURL(*f.rootURL)
	var links map[string][]httputil.Link
	var err error
	client := urlutil.NewClient(5 * time.Second)

	// The flyte takes time to start-up; therefore, we attempt to connect several times
	for i := 0; i < 5; i++ {
		time.Sleep(2 * time.Second)

		if err = client.GetStruct(baseURL.String(), &links); err == nil {
			logger.Infof("Flyte api started at %s", f.rootURL)
			return links, nil
		}
	}
	return nil, fmt.Errorf("Could not get base url: %s\n", err)
}

func (f *Flyte) setPort() error {
	ln, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}
	defer ln.Close()

	port := ln.Addr().(*net.TCPAddr).Port
	if err != nil {
		return err
	}

	f.port = fmt.Sprintf("%d", port)
	f.rootURL, err = url.Parse(fmt.Sprintf("https://localhost:%s", f.port))
	if err != nil {
		return err
	}

	return nil
}

func (f *Flyte) discoverURLs(links map[string][]httputil.Link) error {

	var err error
	f.flowsURL, err = urlutil.FindURLByRel(links["links"], "flow/listFlows")
	if err != nil {
		return err
	}

	f.packsURL, err = urlutil.FindURLByRel(links["links"], "pack/listPacks")
	if err != nil {
		return err
	}

	f.healthURL, err = urlutil.FindURLByRel(links["links"], "info/health")
	if err != nil {
		return err
	}

	f.datastoreURL, err = urlutil.FindURLByRel(links["links"], "datastore/listDataItems")
	if err != nil {
		return err
	}

	f.flowExecutionsURL, err = urlutil.FindURLByRel(links["links"], "audit/findFlows")
	if err != nil {
		return err
	}

	f.swaggerURL, err = urlutil.FindURLByRel(links["links"], "swagger")
	if err != nil {
		return err
	}

	return nil
}

func getBaseURL(u url.URL) url.URL {
	u.Path = "v1"
	return u
}

func (f Flyte) Stop() {
	if f.server != nil {
		f.server.Close()
	}
	f.deleteCertAndKeyFile()
}

func (f Flyte) RootURL() *url.URL {
	return f.rootURL
}

func (f Flyte) PacksURL() string {
	return f.packsURL
}

func (f Flyte) FlowsURL() string {
	return f.flowsURL
}

func (f Flyte) HealthURL() string {
	return f.healthURL
}

func (f Flyte) DatastoreURL() string {
	return f.datastoreURL
}

func (f Flyte) FlowExecutionsURL() string {
	return f.flowExecutionsURL
}

func (f Flyte) SwaggerURL() string {
	return f.swaggerURL
}

var flyteapiKey = `-----BEGIN RSA PRIVATE KEY-----
MIICXAIBAAKBgQCgRgCmgkpl3Xqi9mjEUGZEo/3hkYVHPFbxMQBQpgCisxTTYr3o
pSHid8tIiajNGle5S6HSUFTft/RI4Ninv+d1aaglJF8aAetOiONo1VXtrrdJoxoC
wvKYXv7nTFRu7LkQFD7u5gKxmzNGxrRQyK4UtWXHNvFMTcFQ4G/5bNoDKwIDAQAB
AoGBAINaO8g7OdwYUxzh0+Uoe1hACp9mkxNZyWtsnHR8SEMPf77qyvey9n1TboAp
ifVmZITRDnN+MMEVUxizZfy9U2RjrSougi0ZOv5eC+hXLTSbji3WiKhKDmCBCwEV
GDj/vrJZnAfdGiEfnmXBYpb5oPZBY27P1nSWcmQ79/SZ2jl5AkEAxWtpkO7tE+Q1
Aj6Xl700CK+t2u5+gOUq/EzAB6fidth1j1tHhf+aaZ6H68QowRlSbMZ8jwU/eQtc
UsoI8gfmDQJBAM/U25Zt/8vWyAmGEYSRdSsQ97IUsSD8XcAPPg4jG9Hht+uuX9oi
CDILr3wwLf70WLqJfWah+c8sJ68VIOQ3uBcCQHcotyZQ4G5CLzC0oQFopTCdAT4E
9/xK1qBEnx+/2LRNQOAPg2NA/W3Ez1uiIcszwol/YI1e6IniLo6V/cJAvD0CQHsE
e5XnNmnpkC5S9TuK/deoC3WVWeM0fimY3BpyHZ12Be+zH3l2e3NkB1NzEUbAS2Te
zSNa7Qr8D+FKmFV9xbECQAUW12r6f+0D6hl9pviaTfdidkO2SPsTa+M/pF/8eXiE
IJn7Rr0u4KclZJBXsUAf0C0DJtGx9gXzImInDBkbfbg=
-----END RSA PRIVATE KEY-----`

var flyteapiCert = `-----BEGIN CERTIFICATE-----
MIIB2TCCAUKgAwIBAgIBADANBgkqhkiG9w0BAQsFADAgMQ4wDAYDVQQKEwVmbHl0
ZTEOMAwGA1UEAxMFZmx5dGUwIBcNOTkxMjMxMjM1OTU5WhgPOTk5OTEyMzEyMzU5
NTlaMCAxDjAMBgNVBAoTBWZseXRlMQ4wDAYDVQQDEwVmbHl0ZTCBnzANBgkqhkiG
9w0BAQEFAAOBjQAwgYkCgYEAoEYApoJKZd16ovZoxFBmRKP94ZGFRzxW8TEAUKYA
orMU02K96KUh4nfLSImozRpXuUuh0lBU37f0SODYp7/ndWmoJSRfGgHrTojjaNVV
7a63SaMaAsLymF7+50xUbuy5EBQ+7uYCsZszRsa0UMiuFLVlxzbxTE3BUOBv+Wza
AysCAwEAAaMhMB8wDgYDVR0PAQH/BAQDAgWgMA0GA1UdDgQGBAQBAgMEMA0GCSqG
SIb3DQEBCwUAA4GBAGfj42IqF3pzjzJnv/2yWhBugBg8QKbIQl6IK0jXsbibZ/xJ
oex6pDtwW30StehLkyJKtj5Vq/NSZ3QmRWni9cmzbNiNP6vAyFuATClD+7k6zqh1
zv8h2qSENDzuFRMbyMQVsTPHq250UrsQjJlStAXg/gznXWICbibeWmH2wG0E
-----END CERTIFICATE-----`
