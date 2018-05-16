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

package datastore

import (
	"bytes"
	encodingjson "encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"github.com/HotelsDotCom/flyte/httputil"
	"github.com/HotelsDotCom/go-logger/loggertest"
	"testing"
	"github.com/husobee/vestigo"
	"github.com/HotelsDotCom/flyte/flytepath"
)

func TestGetItems(t *testing.T) {

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		findAll: func() ([]DataItem, error) {
			return []DataItem{{Key: "Item-1", Description: "this is test description"}, {Key: "Item-2"}}, nil
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/datastore", nil)
	httputil.SetProtocolAndHostIn(request)
	w := httptest.NewRecorder()
	GetItems(w, request)
	response := w.Result()

	body, err := ioutil.ReadAll(w.Body)
	require.NoError(t, err, "cannot read response body: %v")
	responseBody := dataItemsResponse{}
	encodingjson.Unmarshal(body, &responseBody)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, httputil.ContentTypeJson, response.Header.Get(httputil.HeaderContentType))
	assert.Equal(t, 2, len(responseBody.DataItems))

	assert.Equal(t, "Item-1", responseBody.DataItems[0].Key)
	assert.Equal(t, "this is test description", responseBody.DataItems[0].Description)
	assert.Equal(t, 1, len(responseBody.DataItems[0].Links))
	assert.Equal(t, "http://example.com/v1/datastore/Item-1", responseBody.DataItems[0].Links[0].Href)
	assert.Equal(t, "self", responseBody.DataItems[0].Links[0].Rel)

	assert.Equal(t, "Item-2", responseBody.DataItems[1].Key)
	assert.Empty(t, responseBody.DataItems[1].Description)
	assert.Equal(t, 1, len(responseBody.DataItems[1].Links))
	assert.Equal(t, "http://example.com/v1/datastore/Item-2", responseBody.DataItems[1].Links[0].Href)
	assert.Equal(t, "self", responseBody.DataItems[1].Links[0].Rel)
}

func TestGetItems_ResponseWithEmptyItems(t *testing.T) {

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		findAll: func() ([]DataItem, error) {
			return []DataItem{}, nil
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/datastore", nil)
	w := httptest.NewRecorder()
	GetItems(w, request)
	response := w.Result()

	body, err := ioutil.ReadAll(w.Body)
	require.NoError(t, err, "cannot read response body: %v")
	responseBody := dataItemsResponse{}
	encodingjson.Unmarshal(body, &responseBody)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, 0, len(responseBody.DataItems))
	assert.NotEqual(t, 0, len(responseBody.Links))
}

func TestGetItems_ServiceError(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		findAll: func() ([]DataItem, error) {
			return nil, errors.New("something went wrong")
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/datastore", nil)
	w := httptest.NewRecorder()
	GetItems(w, request)

	response := w.Result()
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "cannot retrieve data items: something went wrong", logMessages[0].Message)
}

func TestGetItem(t *testing.T) {

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		get: func(key string) (*DataItem, error) {
			if key == "noob" {
				return &DataItem{ContentType: "text/xml", Value: []byte(`<carl>fooks</carl>`)}, nil
			}
			return nil, dataItemNotFound
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/datastore?:key=noob", nil)
	w := httptest.NewRecorder()
	GetItem(w, request)
	response := w.Result()

	body, err := ioutil.ReadAll(w.Body)
	require.NoError(t, err)

	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, "text/xml", response.Header.Get(httputil.HeaderContentType))
	assert.Equal(t, "<carl>fooks</carl>", string(body))
}

func TestGetItem_NotFound(t *testing.T) {

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		get: func(string) (*DataItem, error) {
			return nil, dataItemNotFound
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/datastore", nil)
	w := httptest.NewRecorder()
	GetItem(w, request)
	response := w.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestGetItem_ServiceError(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		get: func(string) (*DataItem, error) {
			return nil, errors.New("something went wrong")
		},
	}

	request := httptest.NewRequest(http.MethodGet, "/datastore", nil)
	w := httptest.NewRecorder()
	GetItem(w, request)

	response := w.Result()
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot retrieve data item key=: something went wrong", logMessages[0].Message)
}

func TestDeleteItem(t *testing.T) {

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		remove: func(key string) error {
			if key == "item_to_delete" {
				return nil
			}
			return errors.New("should not get here")
		},
	}

	request := httptest.NewRequest(http.MethodDelete, "/datastore?:key=item_to_delete", nil)

	w := httptest.NewRecorder()
	DeleteItem(w, request)
	response := w.Result()

	assert.Equal(t, http.StatusNoContent, response.StatusCode)
}

func TestDeleteNonExistingItemReturnsNotFoundResponse(t *testing.T) {

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		remove: func(string) error {
			return dataItemNotFound
		},
	}

	request := httptest.NewRequest(http.MethodDelete, "/datastore?:key=item_to_delete", nil)

	w := httptest.NewRecorder()
	DeleteItem(w, request)
	response := w.Result()

	assert.Equal(t, http.StatusNotFound, response.StatusCode)
}

func TestDeleteItemServiceError(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		remove: func(string) error {
			return errors.New("something went wrong")
		},
	}

	request := httptest.NewRequest(http.MethodDelete, "/datastore?:key=item_to_delete", nil)

	w := httptest.NewRecorder()
	DeleteItem(w, request)

	response := w.Result()
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot delete item key=item_to_delete: something went wrong", logMessages[0].Message)
}

func TestGetDataStoreValue_ShouldReturnJsonForJsonContentTypes(t *testing.T) {

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		get: func(key string) (*DataItem, error) {
			item := DataItem{ContentType: httputil.MediaTypeJson, Value: []byte(`{"flyte": "flyte@flyte.com"}`)}
			return &item, nil
		},
	}

	got, err := GetDataStoreValue("json")

	require.NoError(t, err)
	expected := map[string]interface{}{
		"flyte": "flyte@flyte.com",
	}
	assert.Equal(t, expected, got)
}

func TestGetDataStoreValue_ShouldReturnErrorWhenJsonIsInvalid(t *testing.T) {

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		get: func(key string) (*DataItem, error) {
			item := DataItem{ContentType: httputil.MediaTypeJson, Value: []byte(`--- invalid json`)}
			return &item, nil
		},
	}

	_, err := GetDataStoreValue("invalid_json")

	expectedError := `cannot unmarshal datastore item key=invalid_json: invalid character '-' in numeric literal`
	assert.EqualError(t, err, expectedError)
}

func TestGetDataStoreValue_ShouldReturnStringValueForNonJsonContentTypes(t *testing.T) {

	defer resetDatastoreRepo()
	expectedValue := "hello"
	datastoreRepo = mockDatastoreRepo{
		get: func(key string) (*DataItem, error) {
			item := DataItem{ContentType: "plain/text", Value: []byte(expectedValue)}
			return &item, nil
		},
	}

	got, err := GetDataStoreValue("non_json")

	require.NoError(t, err)
	assert.Equal(t, expectedValue, got.(string))
}

func TestGetDataStoreValue_ShouldReturnError_WhenThereIsIssueGettingItem(t *testing.T) {

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		get: func(key string) (*DataItem, error) {
			return nil, errors.New("unexpected error")
		},
	}

	_, err := GetDataStoreValue("error_item")

	assert.EqualError(t, err, "cannot find datastore item key=error_item: unexpected error")
}

func TestPutItem_ShouldCreateNewItem(t *testing.T) {
	defer resetDatastoreRepo()
	var actualItem DataItem
	datastoreRepo = mockDatastoreRepo{
		add: func(item DataItem) error { actualItem = item; return nil },
		has: func(key string) (bool, error) { return false, nil },
	}

	form := testForm()
	req, err := newMultipartRequest(http.MethodPut, "/v1/datastore/new-item", *form)
	require.NoError(t, err)

	resp := serve(req, http.MethodPut, flytepath.DatastoreItemPath, PutItem)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "http://example.com/v1/datastore/new-item", resp.Header.Get("location"))
	expectedItem := DataItem{Key: "new-item", Description: form.fields["description"], Value: form.fileContent, ContentType: form.fileContentType}
	assert.Equal(t, expectedItem, actualItem)
}

func TestPutItem_ShouldUpdateItem(t *testing.T) {
	defer resetDatastoreRepo()
	var actualItem DataItem
	datastoreRepo = mockDatastoreRepo{
		add: func(item DataItem) error { actualItem = item; return nil },
		has: func(key string) (bool, error) { return true, nil },
	}

	form := testForm()
	form.fields = nil
	req, err := newMultipartRequest(http.MethodPut, "/v1/datastore/existing-item", *form)
	require.NoError(t, err)

	resp := serve(req, http.MethodPut, flytepath.DatastoreItemPath, PutItem)

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.Equal(t, "http://example.com/v1/datastore/existing-item", resp.Header.Get("location"))
	expectedItem := DataItem{Key: "existing-item", Value: form.fileContent, ContentType: form.fileContentType}
	assert.Equal(t, expectedItem, actualItem)
}

func TestPutItem_ShouldFailForEmptyDataItem(t *testing.T) {
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	form := testForm()
	form.fileContent = []byte(``)
	req, err := newMultipartRequest(http.MethodPut, "/v1/datastore/empty-file", *form)
	require.NoError(t, err)

	resp := serve(req, http.MethodPut, flytepath.DatastoreItemPath, PutItem)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Error putting data store item: error getting multipart file: file content is empty", logMessages[0].Message)
}

func TestPutItem_ShouldUseDefaultFileContentTypeIfMissing(t *testing.T) {
	defer resetDatastoreRepo()
	var actualItem DataItem
	datastoreRepo = mockDatastoreRepo{
		add: func(item DataItem) error { actualItem = item; return nil },
		has: func(key string) (bool, error) { return false, nil },
	}

	form := testForm()
	form.fileContentType = ""
	req, err := newMultipartRequest(http.MethodPut, "/v1/datastore/file-with-missing-ct", *form)
	require.NoError(t, err)

	resp := serve(req, http.MethodPut, flytepath.DatastoreItemPath, PutItem)

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, "http://example.com/v1/datastore/file-with-missing-ct", resp.Header.Get("location"))
	expectedItem := DataItem{Key: "file-with-missing-ct", Description: form.fields["description"], Value: form.fileContent, ContentType: "text/plain; charset=us-ascii"}
	assert.Equal(t, expectedItem, actualItem)
}

func TestPutItem_ShouldReturn500WhenUnableToAddToRepo(t *testing.T) {
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		add: func(dataItem DataItem) error { return errors.New("something went wrong") },
		has: func(key string) (bool, error) { return false, nil },
	}

	req, err := newMultipartRequest(http.MethodPut, "/v1/datastore/add-error", *testForm())
	require.NoError(t, err)

	resp := serve(req, http.MethodPut, flytepath.DatastoreItemPath, PutItem)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot add item key=add-error: something went wrong", logMessages[0].Message)
}

func TestPutItem_ShouldReturn500WhenUnableToCheckRepo(t *testing.T) {
	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		has: func(key string) (bool, error) { return false, errors.New("something went wrong") },
	}

	req, err := newMultipartRequest(http.MethodPut, "/v1/datastore/has-error", *testForm())
	require.NoError(t, err)

	resp := serve(req, http.MethodPut, flytepath.DatastoreItemPath, PutItem)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot check if item exists key=has-error: something went wrong", logMessages[0].Message)
}

// --- mocks & helpers ---

type mockDatastoreRepo struct {
	add     func(dataItem DataItem) error
	remove  func(key string) error
	get     func(key string) (*DataItem, error)
	findAll func() ([]DataItem, error)
	has     func(key string) (bool, error)
}

func resetDatastoreRepo() {
	datastoreRepo = datastoreMgoRepo{}
}

func (r mockDatastoreRepo) Add(dataItem DataItem) error {
	return r.add(dataItem)
}

func (r mockDatastoreRepo) Remove(key string) error {
	return r.remove(key)
}

func (r mockDatastoreRepo) Get(key string) (*DataItem, error) {
	return r.get(key)
}

func (r mockDatastoreRepo) FindAll() ([]DataItem, error) {
	return r.findAll()
}

func (r mockDatastoreRepo) Has(key string) (bool, error) {
	return r.has(key)
}

type multipartForm struct {
	fileKey         string
	fileName        string
	fileContentType string
	fileContent     []byte
	fields          map[string]string
}

func newMultipartRequest(method, target string, form multipartForm) (*http.Request, error) {
	body, contentType, err := getMultipartContent(form)
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(method, target, body)
	httputil.SetProtocolAndHostIn(req)
	req.Header.Set(httputil.HeaderContentType, contentType)
	return req, nil
}

func getMultipartContent(form multipartForm) (io.Reader, string, error) {
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)

	if err := writeFormFile(w, form); err != nil {
		return nil, "", err
	}

	for key, val := range form.fields {
		if err := w.WriteField(key, val); err != nil {
			return nil, "", err
		}
	}

	if err := w.Close(); err != nil {
		return nil, "", err
	}

	return body, w.FormDataContentType(), nil
}

func writeFormFile(w *multipart.Writer, form multipartForm) error {
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, form.fileKey, form.fileName))
	h.Set("Content-Type", form.fileContentType)

	part, err := w.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = part.Write(form.fileContent)
	return err
}

func testForm() *multipartForm {
	return &multipartForm{
		fileKey:         "value",
		fileContentType: httputil.MediaTypeJson,
		fileContent:     []byte(`true`),
		fields:          map[string]string{"description": "Data item description"},
	}
}

// Use vestigo router to serve HTTP
// Allows to use unmodified path in the request
// without need to hack it like this `?:my_path_param_key=my_path_param_value`
func serve(req *http.Request, method, path string, handler http.HandlerFunc, middleware ...vestigo.Middleware) *http.Response {
	r := vestigo.NewRouter()
	r.Add(method, path, handler, middleware...)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Result()
}
