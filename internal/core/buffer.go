// Copyright 2017-24 the original author or authors.
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

package core

import (
	"bytes"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

type PooledBuffer struct {
	*bytes.Buffer
}

func NewPooledBuffer() *PooledBuffer {
	return &PooledBuffer{Buffer: bufferPool.Get().(*bytes.Buffer)}
}

func (b *PooledBuffer) Close() error {
	b.Reset()
	bufferPool.Put(b.Buffer)
	return nil
}