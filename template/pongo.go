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

package template

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ExpediaGroup/flyte/datastore"
	"github.com/HotelsDotCom/cronexpr"
	"github.com/flosch/pongo2"
	"math/rand"
	"reflect"
	"regexp"
	"strings"
	"time"
)

var whitespaceRegex = regexp.MustCompile(`\s+`)

func init() {
	pongo2.RegisterFilter("key", getValueByKey)
	pongo2.RegisterFilter("match", match)
	pongo2.RegisterFilter("kvp", keyValuePair)
	pongo2.RegisterFilter("index", index)
	pongo2.RegisterFilter("matchesCron", matchesCron)
	pongo2.RegisterFilter("removedupwhitespaces", removeDupWhitespaces)
	pongo2.RegisterFilter("safecopypaste", safeCopyPaste)
	pongo2.RegisterFilter("extractMatch", extractMatch)

	rand.Seed(time.Now().UTC().UnixNano())
	AddStaticContextEntry("randomInt", randomInt)
	AddStaticContextEntry("randomAlpha", randomAlpha)
	AddStaticContextEntry("base64Encode", base64Encode)
	AddStaticContextEntry("base64Decode", base64Decode)
	AddStaticContextEntry("datastore", datastoreFn)
	AddStaticContextEntry("template", template)
	AddStaticContextEntry("unmarshalJson", unmarshalJson)
}

// Executes a template with given context and returns the rendered template as a string
func execute(template string, context Context) (string, error) {
	templateWithAutoescapeOff := "{% autoescape off %}" + template + "{% endautoescape %}"
	tpl, err := pongo2.FromString(templateWithAutoescapeOff)
	if err != nil {
		return "", err
	}
	if context == nil {
		context = Context{}
	}
	return tpl.Execute(pongo2.Context(context).Update(pongo2.Context(staticContext)))
}

func getValueByKey(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	key := reflect.ValueOf(param.String())
	if !key.IsValid() || !reflect.ValueOf(in.Interface()).IsValid() || !reflect.ValueOf(in.Interface()).MapIndex(key).IsValid() {
		return pongo2.AsValue(nil), nil
	}
	value := reflect.ValueOf(in.Interface()).MapIndex(key)
	return pongo2.AsValue(value.Interface()), nil
}

func match(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	match, _ := regexp.MatchString(param.String(), in.String())
	return pongo2.AsValue(match), nil
}

func extractMatch(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	matcher, _ := regexp.Compile(param.String())
	match := matcher.FindStringSubmatch(in.String())
	if len(match) > 1 {
		return pongo2.AsValue(match[1]), nil
	} else {
		return pongo2.AsValue(in.String()), nil
	}
}

func keyValuePair(in *pongo2.Value, _ *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	result := make(map[string]interface{})
	for _, kvp := range strings.Split(in.String(), ",") {
		if kv := strings.Split(kvp, "="); len(kv) == 2 {
			result[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
		}
	}
	return pongo2.AsValue(result), nil
}

func index(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	if in.CanSlice() && in.Len() > 0 {
		return in.Index(param.Integer()), nil
	}
	return pongo2.AsValue(""), nil
}

func removeDupWhitespaces(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	s := whitespaceRegex.ReplaceAllString(in.String(), " ")
	return pongo2.AsValue(strings.TrimRight(s, " ")), nil
}

func safeCopyPaste(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	replaceNonBreakingSpace := strings.Replace(in.String(), "\u00A0", " ", -1)
	return pongo2.AsValue(replaceNonBreakingSpace), nil
}

func randomInt(upperBound int) int {
	return randomizer(upperBound)
}

var randomizer = func(upperBound int) int {
	return rand.Intn(upperBound)
}

const charset = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ`

func randomAlpha(length int) string {
	if length < 0 {
		panic("word length must be non-negative")
	}
	val := make([]byte, length)
	for i := range val {
		val[i] = charset[rand.Intn(len(charset))]
	}
	return string(val)
}

func template(template *pongo2.Value, context map[string]interface{}) string {

	out, err := execute(template.String(), context)
	if err != nil {
		panic(fmt.Sprintf("resolve: cannot resolve template: %v", err))
	}
	return out
}

func base64Encode(in string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(in))
}

func base64Decode(in string) string {
	out, err := base64.RawURLEncoding.DecodeString(in)
	if err != nil {
		panic(err)
	}
	return string(out)
}

func datastoreFn(key string) interface{} {
	v, err := datastore.GetDataStoreValue(key)
	if err != nil {
		panic(err)
	}
	return v
}

func matchesCron(in *pongo2.Value, param *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
	t, err := time.Parse(time.RFC3339, in.String())
	if err != nil {
		panic("invalid date: '" + in.String() + "'")
	}

	ce, err := cronexpr.Parse(param.String())
	if err != nil {
		return pongo2.AsValue(false), &pongo2.Error{OrigError: err}
	}

	return pongo2.AsValue(ce.Matches(t)), nil
}

func unmarshalJson(in string) map[string]interface{} {
	out := make(map[string]interface{})
	if err := json.Unmarshal([]byte(in), &out); err != nil {
		panic(err)
	}
	return out
}
