// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package encoder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"m4o.io/pbf/v2/model"
)

func TestCoalesceKeepsCallerOrder(t *testing.T) {
	t.Parallel()

	n := func(id model.ID) model.Entity { return &model.Node{ID: id} }
	w := func(id model.ID) model.Entity { return &model.Way{ID: id} }
	r := func(id model.ID) model.Entity { return &model.Relation{ID: id} }

	tests := []struct {
		name string
		in   [][]model.Entity
		size int
		want [][]model.Entity
	}{
		{
			name: "empty",
			in:   nil,
			size: 3,
			want: nil,
		},
		{
			name: "sorted by type",
			in:   [][]model.Entity{{n(1), n(2), w(1)}, {w(2), r(1)}},
			size: 3,
			want: [][]model.Entity{{n(1), n(2), w(1)}, {w(2), r(1)}},
		},
		{
			name: "type changes do not start a batch",
			in:   [][]model.Entity{{r(1), n(1), w(1), n(2), n(3)}},
			size: 3,
			want: [][]model.Entity{{r(1), n(1), w(1)}, {n(2), n(3)}},
		},
		{
			name: "nil entity is dropped",
			in:   [][]model.Entity{{nil, n(1), nil, n(2)}, {nil}},
			size: 3,
			want: [][]model.Entity{{n(1), n(2)}},
		},
		{
			name: "full batch starts a batch",
			in:   [][]model.Entity{{n(1), n(2)}, {n(3)}, {n(4), w(1)}},
			size: 2,
			want: [][]model.Entity{{n(1), n(2)}, {n(3), n(4)}, {w(1)}},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			in := make(chan []model.Entity, len(tc.in))
			for _, entities := range tc.in {
				in <- entities
			}

			close(in)

			var got [][]model.Entity

			for batch := range Coalesce(in, tc.size) {
				require.NoError(t, batch.Error)

				got = append(got, batch.Value)
			}

			assert.Equal(t, tc.want, got)
		})
	}
}
