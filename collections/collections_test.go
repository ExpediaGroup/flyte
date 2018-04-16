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

package collections

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestContainsAll(t *testing.T) {
	testCases := []struct {
		testMap       map[string]string
		shouldContain map[string]string
	}{
		{map[string]string{"env": "staging", "network": "lab"}, map[string]string{"env": "staging", "network": "lab"}},
		{map[string]string{"env": "staging", "network": "lab"}, map[string]string{"network": "lab"}},
		{map[string]string{"env": "staging", "network": "lab"}, map[string]string{"env": "staging"}},
		{map[string]string{"env": "staging", "network": "lab"}, map[string]string{}},
		{map[string]string{"env": "staging", "network": "lab"}, nil},
		{map[string]string{}, map[string]string{}},
		{nil, nil},
	}

	for _, tc := range testCases {
		assert.True(t, ContainsAll(tc.testMap, tc.shouldContain))
	}
}

func TestDoesNotContainAll(t *testing.T) {
	testCases := []struct {
		testMap          map[string]string
		shouldNotContain map[string]string
	}{
		{map[string]string{"env": "staging", "network": "lab"}, map[string]string{"env": "staging", "network": "lab", "foo": "bar"}},
		{map[string]string{"env": "staging", "network": "lab"}, map[string]string{"env": "staging", "network": "prod"}},
		{map[string]string{"env": "staging", "network": "lab"}, map[string]string{"env": "staging", "networkXXX": "prod"}},
		{map[string]string{"env": "staging"}, map[string]string{"env": "staging", "network": ""}},
		{map[string]string{}, map[string]string{"env": "staging"}},
	}

	for _, tc := range testCases {
		assert.False(t, ContainsAll(tc.testMap, tc.shouldNotContain))
	}
}

func TestContains(t *testing.T) {
	s := []string{"alpha", "beta", "gamma", "delta"}
	shouldContain := []string{
		"alpha",
		"delta",
		"gamma",
		"beta",
	}

	for _, c := range shouldContain {
		assert.True(t, Contains(s, c))
	}
}

func TestDoesNotContain(t *testing.T) {
	s := []string{"alpha", "beta", "gamma", "delta"}
	shouldContain := []string{
		"alphaX",
		"foo",
		"",
		"    ",
	}

	for _, c := range shouldContain {
		assert.False(t, Contains(s, c))
	}
}

func TestMergeInCaseOfDuplicateKeysUseValueFromLastMap(t *testing.T) {
	map1 := map[string]string{
		"duplicateKey": "valueFromMap1",
		"key1":         "value1",
	}

	map2 := map[string]string{
		"duplicateKey": "valueFromMap2",
		"key2":         "value2",
		"key3":         "value3",
	}

	want := map[string]string{
		"duplicateKey": "valueFromMap2",
		"key1":         "value1",
		"key2":         "value2",
		"key3":         "value3",
	}
	got := Merge(map1, map2)

	assert.Equal(t, want, got)
}

func TestSortedKeys(t *testing.T) {
	input := map[string]string{
		"bb": "value",
		"b":  "value",
		"a":  "value",
		"c":  "value",
	}
	got := SortedKeys(input)

	want := []string{"a", "b", "bb", "c"}
	assert.Equal(t, want, got)
}

func TestNilSliceDoesNotContain(t *testing.T) {
	assert.False(t, Contains(nil, "foo"))
}

func TestHasMatchingElement(t *testing.T) {
	s1 := []string{"alpha", "beta", "gamma", "delta"}
	shouldHaveMatchingElement := [][]string{
		{"epsilon", "alpha", "omega"},
		{"delta"},
		{"gamma", "beta"},
		{"phi", "beta"},
	}

	for _, s2 := range shouldHaveMatchingElement {
		assert.True(t, HasMatchingElement(s1, s2))
	}
}

func TestDoesNotHaveMatchingElement(t *testing.T) {
	s1 := []string{"alpha", "beta", "gamma", "delta"}
	noMatchingElement := [][]string{
		{"omega"},
		{"phi", "epsilon"},
		{""},
		nil,
	}

	for _, s2 := range noMatchingElement {
		assert.False(t, HasMatchingElement(s1, s2))
	}
}

func TestToStringSlice(t *testing.T) {
	s := []interface{}{"alpha", "beta", "gamma", "delta"}
	actual, err := ToStringSlice(s)
	require.NoError(t, err)

	expected := []string{"alpha", "beta", "gamma", "delta"}
	assert.Equal(t, expected, actual)
}

func TestToStringSliceOnlyWorksWithStrings(t *testing.T) {
	data := []struct {
		toConvert []interface{}
		errorMsg  string
	}{
		{
			toConvert: []interface{}{1, 2, 3, 4},
			errorMsg:  "can only convert slices of type 'string' - slices of type int are not supported",
		},
		{
			toConvert: []interface{}{true, false, true},
			errorMsg:  "can only convert slices of type 'string' - slices of type bool are not supported",
		},
		{
			toConvert: []interface{}{nil, "hello", nil},
			errorMsg:  "can only convert slices of type 'string' - slices of type <nil> are not supported",
		},
	}

	for _, v := range data {
		_, err := ToStringSlice(v.toConvert)
		require.Error(t, err)
		assert.Equal(t, v.errorMsg, err.Error())
	}
}

func TestToStringSliceWithNilSlice(t *testing.T) {
	actual, err := ToStringSlice(nil)
	require.NoError(t, err)
	assert.Equal(t, []string{}, actual)
}
