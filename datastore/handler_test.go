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

func TestPostDataItem(t *testing.T) {

	defer resetDatastoreRepo()
	actualItem := DataItem{}
	datastoreRepo = mockDatastoreRepo{
		add: func(dataItem DataItem) error {
			actualItem = dataItem
			return nil
		},
	}

	form := map[string]string{"key": "itemABC", "description": "test item ABC"}
	value := []byte(`{"one": "1", "two", "2"}`)
	header, body := multipartPostRequest(t, form, value, httputil.MediaTypeJson)
	request := httptest.NewRequest(http.MethodPost, "/datastore", body)
	request.Header = header
	httputil.SetProtocolAndHostIn(request)

	w := httptest.NewRecorder()
	PostItem(w, request)
	response := w.Result()

	assert.Equal(t, http.StatusCreated, response.StatusCode)
	expectedItem := DataItem{Key: "itemABC", Description: "test item ABC", Value: value, ContentType: httputil.MediaTypeJson}
	assert.Equal(t, expectedItem, actualItem)
	assert.Equal(t, "http://example.com/v1/datastore/itemABC", response.Header.Get("location"))
}

func TestPostEmptyDataItem(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	form := map[string]string{"key": "empty_item"}
	value := []byte("")
	header, body := multipartPostRequest(t, form, value, httputil.MediaTypeJson)
	request := httptest.NewRequest(http.MethodPost, "/datastore", body)
	request.Header = header

	w := httptest.NewRecorder()
	PostItem(w, request)

	response := w.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Error posting data store item: error getting multipart file: file content is empty", logMessages[0].Message)
}

func TestPostDataItemWithMissingRequiredFieldsShouldFail(t *testing.T) {

	var cases = []struct {
		form  map[string]string
		value string
	}{
		{form: map[string]string{"key": "the key"}, value: ``},
		{form: map[string]string{}, value: `{"json": "value"}`},
	}

	for _, c := range cases {

		header, body := multipartPostRequest(t, c.form, []byte(c.value), httputil.MediaTypeJson)
		request := httptest.NewRequest(http.MethodPost, "/datastore", body)
		request.Header = header

		w := httptest.NewRecorder()
		PostItem(w, request)
		response := w.Result()

		assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	}
}

func TestPostDataItemMissingMultipartHeaderDefaultValue(t *testing.T) {

	defer resetDatastoreRepo()
	actualItem := DataItem{}
	datastoreRepo = mockDatastoreRepo{
		add: func(dataItem DataItem) error {
			actualItem = dataItem
			return nil
		},
	}

	form := map[string]string{"key": "ItemWithMissingContentType"}
	value := []byte(`[1, 2, 3]`)
	header, body := multipartPostRequest(t, form, value, "")
	request := httptest.NewRequest(http.MethodPost, "/datastore", body)
	request.Header = header
	httputil.SetProtocolAndHostIn(request)

	w := httptest.NewRecorder()
	PostItem(w, request)
	response := w.Result()

	assert.Equal(t, http.StatusCreated, response.StatusCode)
	expectedItem := DataItem{Key: "ItemWithMissingContentType", Value: value, ContentType: "text/plain; charset=us-ascii"}
	assert.Equal(t, expectedItem, actualItem)
	assert.Empty(t, actualItem.Description)
	assert.Equal(t, "http://example.com/v1/datastore/ItemWithMissingContentType", response.Header.Get("location"))
}

func TestPostDataItemServiceFailed(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		add: func(dataItem DataItem) error {
			return errors.New("something went wrong")
		},
	}

	form := map[string]string{"key": "failing_item"}
	value := []byte(`some value`)
	header, body := multipartPostRequest(t, form, value, "plain/text")
	request := httptest.NewRequest(http.MethodPost, "/datastore", body)
	request.Header = header

	w := httptest.NewRecorder()
	PostItem(w, request)

	response := w.Result()
	assert.Equal(t, http.StatusInternalServerError, response.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot add item key=failing_item: something went wrong", logMessages[0].Message)
}

func TestPostDuplicateDataItem(t *testing.T) {

	defer loggertest.Reset()
	loggertest.Init(loggertest.LogLevelError)

	defer resetDatastoreRepo()
	datastoreRepo = mockDatastoreRepo{
		add: func(dataItem DataItem) error {
			return dataItemExists
		},
	}

	form := map[string]string{"key": "duplicate_item"}
	value := []byte(`---`)
	header, body := multipartPostRequest(t, form, value, "text/plain")
	request := httptest.NewRequest(http.MethodPost, "/datastore", body)
	request.Header = header

	w := httptest.NewRecorder()
	PostItem(w, request)

	response := w.Result()
	assert.Equal(t, http.StatusConflict, response.StatusCode)

	logMessages := loggertest.GetLogMessages()
	require.Len(t, logMessages, 1)
	assert.Equal(t, "Cannot add item, key=duplicate_item already exists", logMessages[0].Message)
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

// --- mocks & helpers ---

func multipartPostRequest(t *testing.T, form map[string]string, fileContent []byte, fileContentType string) (http.Header, io.Reader) {

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)
	writer.FormDataContentType()

	if err := writeMultipartFile(writer, fileContent, fileContentType); err != nil {
		assert.Fail(t, fmt.Sprintf("cannot create form file: %v", err))
	}

	for k, v := range form {
		if err := writer.WriteField(k, v); err != nil {
			assert.Fail(t, fmt.Sprintf("cannot write form field: %v", err))
		}
	}

	if err := writer.Close(); err != nil {
		assert.Fail(t, fmt.Sprintf("error closing multipart writer: %v", err))
	}

	return http.Header{httputil.HeaderContentType: []string{writer.FormDataContentType()}}, body
}

func writeMultipartFile(writer *multipart.Writer, content []byte, contentType string) error {

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="value"; filename="test_file_item"`)
	h.Set(httputil.HeaderContentType, contentType)
	part, err := writer.CreatePart(h)
	if err != nil {
		return err
	}

	_, err = part.Write(content)
	return err
}

type mockDatastoreRepo struct {
	add     func(dataItem DataItem) error
	remove  func(key string) error
	get     func(key string) (*DataItem, error)
	findAll func() ([]DataItem, error)
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
