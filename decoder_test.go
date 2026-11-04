// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package pbf

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m4o.io/pbf/v2/internal/encoder"
	"m4o.io/pbf/v2/model"
)

func TestDecodeSample(t *testing.T) {
	publicDecodeOsmPbf(t, "testdata/sample.osm.pbf", 339)
}

func TestNewDecoderFailsOnUnknownRequiredFeature(t *testing.T) {
	var encoded bytes.Buffer

	enc, err := NewEncoder(&encoded, WithRequiredFeatures("Unknown-Feature-X"))
	if err != nil {
		t.Fatalf("create encoder: %v", err)
	}
	enc.Close()

	dec, err := NewDecoder(context.Background(), bytes.NewReader(encoded.Bytes()))
	if err == nil {
		dec.Close()
		t.Fatal("expected NewDecoder to fail on unknown required feature")
	}
	if !errors.Is(err, ErrUnsupportedRequiredFeature) {
		t.Fatalf("expected ErrUnsupportedRequiredFeature, got: %v", err)
	}
}

func publicDecodeOsmPbf(t *testing.T, file string, expectedEntries int) {
	in, err := os.Open(file)
	if err != nil {
		t.Errorf("Error reading file: %v", err)
	}

	defer in.Close()

	// decode header blob
	decoder, err := NewDecoder(context.Background(), in)
	require.NoError(t, err)

	assert.Equal(t, reflect.TypeOf(model.Header{}), reflect.TypeOf(decoder.Header))

	// decode entities
	var nEntries int

	for {
		objs, err := decoder.Decode()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				t.Errorf("Error decoding%v", err)
			} else {
				break
			}
		}

		for _, obj := range objs {
			assert.NotEqual(t, reflect.TypeOf(model.Header{}), reflect.TypeOf(obj))
		}

		nEntries += len(objs)
	}

	assert.Equal(t, expectedEntries, nEntries, "Incorrect number of entities")
}

func TestDecoderKeepsFileOrder(t *testing.T) {
	t.Parallel()

	const nBlocks = 32

	nEntities := nBlocks * encoder.EntityLimit

	var encoded bytes.Buffer

	enc, err := NewEncoder(&encoded)
	require.NoError(t, err)

	for id := range nEntities {
		node := &model.Node{ID: model.ID(id), Info: &model.Info{Visible: true}}
		require.NoError(t, enc.Encode(node))
	}

	enc.Close()

	dec, err := NewDecoder(context.Background(), bytes.NewReader(encoded.Bytes()), WithNCpus(4), WithProtoBatchSize(1))
	require.NoError(t, err)

	defer dec.Close()

	var got []model.ID

	for {
		batch, err := dec.Decode()
		if errors.Is(err, io.EOF) {
			break
		}

		require.NoError(t, err)

		for _, entity := range batch {
			got = append(got, entity.GetID())
		}
	}

	require.Len(t, got, nEntities)

	for i, id := range got {
		require.Equal(t, model.ID(i), id, "entity %d", i)
	}
}
