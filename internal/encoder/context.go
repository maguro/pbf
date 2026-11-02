// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

import "io"

type Context struct {
	wrtr        io.Writer
	compression BlobCompression
}

func NewContext(wrtr io.Writer, compression BlobCompression) *Context {
	return &Context{
		wrtr:        wrtr,
		compression: compression,
	}
}
