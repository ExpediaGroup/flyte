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

package auth

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClaimsFulfilled(t *testing.T) {
	tokenClaims := jwt.MapClaims{
		"foo":            "bar",
		"user":           "Mr X",
		"superuser":      true,
		"role":           "admin",
		"accesslevel":    1,
		"groups":         []interface{}{"dev", "role"},
		"groups2":        []string{"dev2", "role2"},
		"somebools":      []bool{true, false, true},
		"willgetignored": []interface{}{1.2, 1.3, 1.4},
	}

	policyClaimsSlice := []policyClaims{
		{
			"groups": {"foo", "bar"},
			"role":   {"admin"},
		},
		{
			"groups": {"foo", "bar", "dev"},
		},
		{
			"groups2": {"dev2"},
		},
		{
			"role": {"admin"},
		},
		{
			"superuser": {"true"},
			"groups":    {"blah", "blah blah"},
		},
		{
			"accesslevel": {"1"},
		},
		{},
		nil,
	}

	for _, p := range policyClaimsSlice {
		assert.True(t, p.fulfilled(tokenClaims, nil))
	}
}

func TestClaimsNotFulfilled(t *testing.T) {
	tokenClaims := jwt.MapClaims{
		"foo":         "bar",
		"user":        "Mr X",
		"superuser":   true,
		"role":        "admin",
		"accesslevel": 1,
		"groups":      []interface{}{"dev", "role"},
		"groups2":     []string{"dev2", "role2"},
		"somebools":   []bool{true, false, true},
	}

	policyClaimsSlice := []policyClaims{
		{
			"groups": {"foo", "bar"},
			"role":   {"superuser"},
		},
		{
			"groups": {"foo", "bar"},
		},
		{
			// []string (or []interface{} that can be mapped to []string) are the only slice types that can be matched in a token
			// - here we are trying to match []bool
			"somebools": {"false"},
		},
		{
			"superuser": {"false"},
		},
		{
			"accesslevel": {"2"},
		},
	}

	for _, p := range policyClaimsSlice {
		assert.False(t, p.fulfilled(tokenClaims, nil))
	}
}

func TestDynamicClaimsFulfilled(t *testing.T) {
	tokenClaims := jwt.MapClaims{
		"foo":         "bar",
		"user":        "Mr X",
		"superuser":   true,
		"role":        "admin",
		"accesslevel": 1,
		"groups":      []interface{}{"dev", "role"},
		"groups2":     []string{"dev2", "role2"},
		"somebools":   []bool{true, false, true},
		"pack":        "bamboo",
		"namespace":   "com.some.namespace",
	}

	placeholderVals := map[string]string{
		"pack":          "bamboo",
		"namespace":     "com.some.namespace",
		"accesscontrol": "groups",
	}

	policyClaimsSlice := []policyClaims{
		// based on placeholderVals resolves to "pack" : {"bamboo"}
		{
			"pack": {":pack"},
		},
		// based on placeholderVals resolves to "namespace" : {"com.some.namespace"}
		{
			"namespace": {":namespace"},
		},
		// based on placeholderVals resolves to "groups" : {"dev", "ssp"}
		{
			":accesscontrol": {"dev", "ssp"},
		},
	}

	for _, p := range policyClaimsSlice {
		assert.True(t, p.fulfilled(tokenClaims, placeholderVals))
	}
}

func TestDynamicClaimsNotFulfilled(t *testing.T) {
	tokenClaims := jwt.MapClaims{
		"foo":         "bar",
		"user":        "Mr X",
		"superuser":   true,
		"role":        "admin",
		"accesslevel": 1,
		"groups":      []interface{}{"dev", "role"},
		"groups2":     []string{"dev2", "role2"},
		"somebools":   []bool{true, false, true},
		"pack":        "bamboo",
		"namespace":   "com.some.namespace",
	}

	placeholderVals := map[string]string{
		"pack":          "github",
		"accesscontrol": "groups2",
	}

	policyClaimsSlice := []policyClaims{
		// based on placeholderVals resolves to "pack" : {"github"} which does not match token
		{
			"pack": {":pack"},
		},
		// based on placeholderVals 'resolves' to "namespace" : {""} which does not match token
		{
			"namespace": {":namespace"},
		},
		// based on placeholderVals resolves to "groups2" : {"dev", "ssp"}
		{
			":accesscontrol": {"dev", "ssp"},
		},
	}

	for _, p := range policyClaimsSlice {
		assert.False(t, p.fulfilled(tokenClaims, placeholderVals))
	}
}
