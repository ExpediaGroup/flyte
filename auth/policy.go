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
	"fmt"
	"github.com/ExpediaGroup/flyte/collections"
	"github.com/HotelsDotCom/go-logger"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"strconv"
	"strings"
)

type pathPolicy struct {
	Path        string
	HttpMethods []string `yaml:"methods"`
	Claims      policyClaims
}

func newPathPolicies(policyPath string) ([]pathPolicy, error) {
	b, err := ioutil.ReadFile(policyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load auth policy file %q: %v", policyPath, err)
	}
	pathPolicies := []pathPolicy{}
	if err := yaml.Unmarshal(b, &pathPolicies); err != nil {
		return nil, fmt.Errorf("failed to deserialise auth policy file %q: %v", policyPath, err)
	}
	return pathPolicies, nil
}

type policyClaims map[string][]string

func (c policyClaims) fulfilled(tokenClaims jwt.MapClaims, placeholderVals map[string]string) bool {
	// having no policy claims means that any valid token satisfies them
	if len(c) == 0 {
		return true
	}
	for policyClaimName, policyClaimVals := range c {

		policyClaimName = resolveTemplate(policyClaimName, placeholderVals)
		policyClaimVals = resolveTemplates(policyClaimVals, placeholderVals)

		if tokenClaimVal, ok := tokenClaims[policyClaimName]; ok {
			switch tClaimVal := tokenClaimVal.(type) {
			case string:
				if collections.Contains(policyClaimVals, tClaimVal) {
					return true
				}
			case bool:
				if collections.Contains(policyClaimVals, strconv.FormatBool(tClaimVal)) {
					return true
				}
			case int:
				if collections.Contains(policyClaimVals, strconv.Itoa(tClaimVal)) {
					return true
				}
			case []string:
				if collections.HasMatchingElement(policyClaimVals, tClaimVal) {
					return true
				}
				// this handles []string that has been unmarshalled/declared as []interface{}
			case []interface{}:
				tClaimStrings, err := collections.ToStringSlice(tClaimVal)
				if err != nil {
					logger.Infof("unsupported jwt claims type: %v", tClaimVal, err)
					continue
				}
				if collections.HasMatchingElement(policyClaimVals, tClaimStrings) {
					return true
				}
			default:
				logger.Infof("jwt claim values of type %T are not supported", tClaimVal)
			}
		}
	}
	return false
}

func resolveTemplates(val []string, placeholderVals map[string]string) []string {
	resolved := make([]string, len(val))
	for i, v := range val {
		resolved[i] = resolveTemplate(v, placeholderVals)
	}
	return resolved
}

func resolveTemplate(val string, placeholderVals map[string]string) string {
	resolved := val
	if strings.HasPrefix(val, ":") {
		resolved = placeholderVals[strings.TrimPrefix(val, ":")]
	}
	return resolved
}
