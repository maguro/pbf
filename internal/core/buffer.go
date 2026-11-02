// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"bytes"
	"sync"
)

const bufferSize = 1024

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, bufferSize))
	},
}

type PooledBuffer struct {
	*bytes.Buffer
}

func NewPooledBuffer() *PooledBuffer {
	//nolint:forcetypeassert // bufferPool holds only *bytes.Buffer values.
	return &PooledBuffer{Buffer: bufferPool.Get().(*bytes.Buffer)}
}

func (b *PooledBuffer) Close() error {
	b.Reset()
	bufferPool.Put(b.Buffer)

	return nil
}
