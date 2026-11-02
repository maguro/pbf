// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"os"

	"github.com/spf13/pflag"
)

// -- *os.File Value.
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
