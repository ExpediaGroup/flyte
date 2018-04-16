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
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"testing"
	"time"
)

type HttpClient struct {
	*http.Client
}

func NewClient(timeout time.Duration) HttpClient {
	return HttpClient{&http.Client{
		Timeout:   timeout,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
	}}
}

func (c HttpClient) PostAndAssert(url, requestBody, bodyContent string, t *testing.T) *http.Response {
	resp, err := c.Post(url, requestBody)
	if err != nil {
		t.Fatalf("Failed to post %s: %s", bodyContent, err)
	}
	return resp
}

func (c HttpClient) Post(url, requestBody string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.Do(req)
}

func (c HttpClient) PostResourceAndAssert(url, requestBody, bodyContent string, t *testing.T) *url.URL {
	location, err := c.PostResource(url, requestBody)
	if err != nil {
		t.Fatalf("Failed to post %s: %s", bodyContent, err)
	}
	return location
}

func (c HttpClient) PostResource(url, requestBody string) (location *url.URL, err error) {
	resp, err := c.Post(url, requestBody)
	if err != nil {
		return nil, err
	}
	return resp.Location()
}

func (c HttpClient) PostStruct(url string, body interface{}) (*http.Response, error) {
	b, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("cannot marshal body '%+v': %v", body, err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.Do(req)
}

func (c HttpClient) PostWithStructResponse(url, requestBody string, response interface{}) error {
	resp, err := c.Post(url, requestBody)
	if err != nil {
		return fmt.Errorf("Failed to post %q to %q: %s", requestBody, url, err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return fmt.Errorf("could not deserialise response from %q: %s", url, err)
	}
	return nil
}

func (c HttpClient) PostMultipart(url string, form map[string]string, fileContent []byte, fileContentType string) (*http.Response, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	if err := writeMultipartFile(writer, fileContent, fileContentType); err != nil {
		return nil, fmt.Errorf("cannot create form file: %v", err)
	}

	for k, v := range form {
		if err := writer.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("cannot write form field: %v", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %v", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	return c.Do(req)
}

func writeMultipartFile(writer *multipart.Writer, fileContent []byte, contentType string) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="value"; filename="test_file_item"`)
	h.Set("Content-Type", contentType)
	part, err := writer.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = part.Write(fileContent)
	return err
}

func (c HttpClient) GetStruct(url string, s interface{}) error {
	resp, err := c.Get(url)
	if err != nil {
		return fmt.Errorf("error getting url %q: %s", url, err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(s)
	if err != nil {
		return fmt.Errorf("could not deserialise response from %q: %s", url, err)
	}
	return nil
}

func (c HttpClient) Get(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %v", err)
	}
	req.Header.Set("Accept", "application/json")

	return c.Do(req)
}

func (c HttpClient) Delete(url string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cannot create request: %v", err)
	}
	return c.Do(req)
}

func (c HttpClient) DeleteResource(t *testing.T, location *url.URL) {
	if _, err := c.Delete(location.String()); err != nil {
		t.Fatalf("Could not delete resource at location %q: %s", location.String(), err)
	}
}
