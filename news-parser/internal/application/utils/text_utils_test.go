package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeText(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "replaces non breaking spaces and trims edges",
			input:    "\u00a0 bitcoin\u00a0price \u00a0",
			expected: "bitcoin price",
		},
		{
			name:     "replaces line breaks with single spaces",
			input:    "first line\r\nsecond line\nthird line",
			expected: "first line second line third line",
		},
		{
			name:     "collapses repeated whitespace",
			input:    "bitcoin    market \n\n sentiment",
			expected: "bitcoin market sentiment",
		},
		{
			name:     "keeps already normalized text intact",
			input:    "clean text",
			expected: "clean text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, NormalizeText(tt.input))
		})
	}
}
