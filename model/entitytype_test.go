// SPDX-FileCopyrightText: the original author or authors
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntityTypeString(t *testing.T) {
	assert.Equal(t, "NODE", NODE.String())
	assert.Equal(t, "WAY", WAY.String())
	assert.Equal(t, "RELATION", RELATION.String())

	assert.Equal(t, "EntityType(-1)", (NODE - 1).String())
	assert.Equal(t, "EntityType(3)", (RELATION + 1).String())
}
