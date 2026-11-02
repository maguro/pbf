// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package packers

import (
	"io"
)

type base struct {
	w io.WriteCloser
}

func newBasePacker(w io.WriteCloser) *base {
	return &base{w: w}
}

func (rp *base) Write(b []byte) (int, error) {
	return rp.w.Write(b)
}

func (rp *base) Close() error {
	return rp.w.Close()
}
