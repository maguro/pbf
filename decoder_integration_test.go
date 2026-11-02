//go:build integration
// +build integration

// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"m4o.io/pbf/v2/model"
)

func TestDecodeBremen(t *testing.T) {
	publicDecodeOsmPbf(t, "testdata/bremen.osm.pbf", 1640420)
}

func TestDecodeLondon(t *testing.T) {
	publicDecodeOsmPbf(t, "testdata/greater-london.osm.pbf", 3200894)
}

func TestDecoderStop(t *testing.T) {
	in, err := os.Open("testdata/greater-london.osm.pbf")
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	defer in.Close()

	// decode header blob
	decoder, err := NewDecoder(context.Background(), in)
	if err != nil {
		t.Errorf("Error reading blob header: %v", err)
	}

	assert.Equal(t, reflect.TypeOf(model.Header{}), reflect.TypeOf(decoder.Header))

	cancel := make(chan bool, 1)

	go func() {
		<-cancel
		decoder.Close()
		decoder.Close()
	}()

	// decode entities
	var nEntries int

	for {
		if nEntries == 1000 {
			cancel <- true
		}

		e, err := decoder.Decode()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Errorf("Error decoding%v", err)
			} else {
				break
			}
		}

		assert.NotEqual(t, reflect.TypeOf(model.Header{}), reflect.TypeOf(e))

		nEntries++
	}

	assert.True(t, nEntries < 3200894, "Close did not cancel decoding")
}
