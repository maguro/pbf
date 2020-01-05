// Copyright 2017-20 the original author or authors.
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

package cli

import (
	"os"

	"github.com/spf13/pflag"
)

// -- *os.File Value
type readerValue struct {
	value    **os.File
	typename string
}

// NewReaderValue creates an cobra Value object for an *os.File.
func NewReaderValue(def *os.File, p **os.File, typename string) pflag.Value {
	bbv := &readerValue{
		value:    p,
		typename: typename,
	}
	*bbv.value = def

	return bbv
}

func (r *readerValue) Set(val string) error {
	f, err := os.Open(val)
	if err != nil {
		return err
	}

	*r.value = f

	return nil
}

func (r *readerValue) Type() string {
	return r.typename
}

func (r *readerValue) String() string {
	if *r.value == nil {
		return ""
	}

	return (*r.value).Name()
}
