// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

import "io"

type Context struct {
	w           io.Writer
	compression BlobCompression
}

func NewContext(w io.Writer, compression BlobCompression) *Context {
	return &Context{
		w:           w,
		compression: compression,
	}
}
