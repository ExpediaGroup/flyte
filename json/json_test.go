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

package json

import (
	encodingjson "encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode"
)

func TestSerialisationDeserialisation(t *testing.T) {

	cases := []string{
		`
		{
			"anArray": [1, 2, 5, 111],
			"foo": "bar",
			"isItTrue": true,
			"obj": {
				"foo2": "bar2",
				"foo3": "bar3"
			}
		}
		`,

		`false`,
		`"a string"`,
		`12345`,
		`12.345`,
		`[1,2,3,4,5]`,
	}

	for i, val := range cases {
		t.Run(fmt.Sprintf("%v", i), func(t *testing.T) { SerialisationDeserialisationTest(val, t) })
	}
}

func SerialisationDeserialisationTest(val string, t *testing.T) {

	j, err := NewJson(strings.NewReader(val))
	if err != nil {
		t.Error("Could not parse json", err)
	}

	jsonBytes, err := encodingjson.Marshal(j)
	if err != nil {
		t.Error("Could not serialise json", err)
	}

	val2 := string(jsonBytes)

	val = removeWhitespace(val)
	val2 = removeWhitespace(val2)
	if val != val2 {
		t.Errorf("Deserialised/serialised json does not match original json. Expected: %s, Actual:%s", val, val2)
	}
}

func removeWhitespace(v string) string {

	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, v)
}
