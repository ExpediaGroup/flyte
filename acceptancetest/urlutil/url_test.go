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
	"testing"
)

func TestJoin(t *testing.T) {

	var cases = []struct {
		path     string
		elements []string
		expected string
	}{
		{"test/", []string{}, "test/"},
		{"test/", []string{"/v1/child"}, "test/v1/child"},
		{"test", []string{"", ""}, "test"},
		{"/a/", []string{}, "/a/"},
		{"/a", []string{"b", "c/"}, "/a/b/c/"},
		{"/a/", []string{"/b/", "c/"}, "/a/b/c/"},
		{"http://www.example.com/", []string{"/a/", "/b/"}, "http://www.example.com/a/b/"},
		{"http://www.example.com/", []string{"/a/", "b/"}, "http://www.example.com/a/b/"},
		{"http://www.example.com/", []string{"/a", "b/"}, "http://www.example.com/a/b/"},
		{"http://www.example.com/", []string{"a", "b"}, "http://www.example.com/a/b"},
		{"http://www.example.com", []string{"/a/", "b/"}, "http://www.example.com/a/b/"},
		{"http://www.example.com", []string{"/a/", "/b/"}, "http://www.example.com/a/b/"},
		{"http://www.example.com", []string{"a/", "b/"}, "http://www.example.com/a/b/"},
		{"http://www.example.com", []string{"a", "b"}, "http://www.example.com/a/b"},
	}

	for _, c := range cases {
		actual := Join(c.path, c.elements...)
		if actual != c.expected {
			t.Errorf("\nJoin(%q %v):\nExpected: %q\nActual:   %q\n", c.path, c.elements, c.expected, actual)
		}
	}
}
