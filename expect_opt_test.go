// Copyright 2018 Netflix, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package expect

import (
	"bytes"
	"errors"
	"io"
	"regexp"
	"testing"
)

func checkMatcher(t *testing.T, m Matcher, expected bool) {
	t.Helper()
	if expected != (m != nil) {
		if m == nil {
			t.Error("got nil matcher, wanted non-nil")
		} else {
			t.Errorf("got matcher %+#v, wanted nil", m)
		}
	}
}

func TestExpectOptString(t *testing.T) {
	tests := []struct {
		title    string
		opt      ExpectOpt
		data     string
		expected bool
	}{
		{
			"No args",
			String(),
			"Hello world",
			false,
		},
		{
			"Single arg",
			String("Hello"),
			"Hello world",
			true,
		},
		{
			"Multiple arg",
			String("other", "world"),
			"Hello world",
			true,
		},
		{
			"No matches",
			String("hello"),
			"Hello world",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			var options ExpectOpts
			err := test.opt(&options)
			checkErr(t, "test.opt", err, nil)

			buf := new(bytes.Buffer)
			_, err = buf.WriteString(test.data)
			checkErr(t, "bytes.Buffer.WriteString", err, nil)

			matcher := options.Match(buf)
			checkMatcher(t, matcher, test.expected)
		})
	}
}

func TestExpectOptRegexp(t *testing.T) {
	tests := []struct {
		title    string
		opt      ExpectOpt
		data     string
		expected bool
	}{
		{
			"No args",
			Regexp(),
			"Hello world",
			false,
		},
		{
			"Single arg",
			Regexp(regexp.MustCompile(`^Hello`)),
			"Hello world",
			true,
		},
		{
			"Multiple arg",
			Regexp(regexp.MustCompile(`^Hello$`), regexp.MustCompile(`world$`)),
			"Hello world",
			true,
		},
		{
			"No matches",
			Regexp(regexp.MustCompile(`^Hello$`)),
			"Hello world",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			var options ExpectOpts
			err := test.opt(&options)
			checkErr(t, "test.opt", err, nil)

			buf := new(bytes.Buffer)
			_, err = buf.WriteString(test.data)
			checkErr(t, "bytes.Buffer.WriteString", err, nil)

			matcher := options.Match(buf)
			checkMatcher(t, matcher, test.expected)
		})
	}
}

func TestExpectOptRegexpPattern(t *testing.T) {
	tests := []struct {
		title    string
		opt      ExpectOpt
		data     string
		expected bool
	}{
		{
			"No args",
			RegexpPattern(),
			"Hello world",
			false,
		},
		{
			"Single arg",
			RegexpPattern(`^Hello`),
			"Hello world",
			true,
		},
		{
			"Multiple arg",
			RegexpPattern(`^Hello$`, `world$`),
			"Hello world",
			true,
		},
		{
			"No matches",
			RegexpPattern(`^Hello$`),
			"Hello world",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			var options ExpectOpts
			err := test.opt(&options)
			checkErr(t, "test.opt", err, nil)

			buf := new(bytes.Buffer)
			_, err = buf.WriteString(test.data)
			checkErr(t, "bytes.Buffer.WriteString", err, nil)

			matcher := options.Match(buf)
			checkMatcher(t, matcher, test.expected)
		})
	}
}

func TestExpectOptError(t *testing.T) {
	tests := []struct {
		title    string
		opt      ExpectOpt
		data     error
		expected bool
	}{
		{
			"No args",
			Error(),
			io.EOF,
			false,
		},
		{
			"Single arg",
			Error(io.EOF),
			io.EOF,
			true,
		},
		{
			"Multiple arg",
			Error(io.ErrShortWrite, io.EOF),
			io.EOF,
			true,
		},
		{
			"No matches",
			Error(io.ErrShortWrite),
			io.EOF,
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			var options ExpectOpts
			err := test.opt(&options)
			checkErr(t, "test.opt", err, nil)

			matcher := options.Match(test.data)
			checkMatcher(t, matcher, test.expected)
		})
	}
}

func TestExpectOptThen(t *testing.T) {
	var (
		errFirst  = errors.New("first")
		errSecond = errors.New("second")
	)

	tests := []struct {
		title    string
		opt      ExpectOpt
		data     string
		match    bool
		expected error
	}{
		{
			"Noop",
			String("Hello").Then(func(buf *bytes.Buffer) error {
				return nil
			}),
			"Hello world",
			true,
			nil,
		},
		{
			"Short circuit",
			String("Hello").Then(func(buf *bytes.Buffer) error {
				return errFirst
			}).Then(func(buf *bytes.Buffer) error {
				return errSecond
			}),
			"Hello world",
			true,
			errFirst,
		},
		{
			"Chain",
			String("Hello").Then(func(buf *bytes.Buffer) error {
				return nil
			}).Then(func(buf *bytes.Buffer) error {
				return errSecond
			}),
			"Hello world",
			true,
			errSecond,
		},
		{
			"No matches",
			String("other").Then(func(buf *bytes.Buffer) error {
				return errFirst
			}),
			"Hello world",
			false,
			nil,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			var options ExpectOpts
			err := test.opt(&options)
			checkErr(t, "test.opt", err, nil)

			buf := new(bytes.Buffer)
			_, err = buf.WriteString(test.data)
			checkErr(t, "bytes.Buffer.WriteString", err, nil)

			matcher := options.Match(buf)
			checkMatcher(t, matcher, test.match)
			if test.match && matcher != nil {
				cb, ok := matcher.(CallbackMatcher)
				if !ok {
					t.Errorf("got %T, wanted an expect.CallbackMatcher", matcher)
				} else {
					err = cb.Callback(nil)
					checkErr(t, "Callback", err, test.expected)
				}
			}
		})
	}
}

func TestExpectOptAll(t *testing.T) {
	tests := []struct {
		title    string
		opt      ExpectOpt
		data     string
		expected bool
	}{
		{
			"No opts",
			All(),
			"Hello world",
			true,
		},
		{
			"Single string match",
			All(String("Hello")),
			"Hello world",
			true,
		},
		{
			"Single string no match",
			All(String("Hello")),
			"No match",
			false,
		},
		{
			"Ordered strings match",
			All(String("Hello"), String("world")),
			"Hello world",
			true,
		},
		{
			"Ordered strings not all match",
			All(String("Hello"), String("world")),
			"Hello",
			false,
		},
		{
			"Unordered strings",
			All(String("world"), String("Hello")),
			"Hello world",
			true,
		},
		{
			"Unordered strings not all match",
			All(String("world"), String("Hello")),
			"Hello",
			false,
		},
		{
			"Repeated strings match",
			All(String("Hello"), String("Hello")),
			"Hello world",
			true,
		},
		{
			"Mixed opts match",
			All(String("Hello"), RegexpPattern(`wo[a-z]{1}ld`)),
			"Hello woxld",
			true,
		},
		{
			"Mixed opts no match",
			All(String("Hello"), RegexpPattern(`wo[a-z]{1}ld`)),
			"Hello wo4ld",
			false,
		},
	}

	for _, test := range tests {
		t.Run(test.title, func(t *testing.T) {
			var options ExpectOpts
			err := test.opt(&options)
			checkErr(t, "test.opt", err, nil)

			buf := new(bytes.Buffer)
			_, err = buf.WriteString(test.data)
			checkErr(t, "bytes.Buffer.WriteString", err, nil)

			matcher := options.Match(buf)
			checkMatcher(t, matcher, test.expected)
		})
	}
}
