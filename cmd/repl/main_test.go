package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringToAnySlice(t *testing.T) {
	assert.Equal(t, []any{"aabc", "abc", "abd"}, stringToAnySlice([]string{"aabc", "abc", "abd"}))
}

func TestFindLongestCommonPrefix(t *testing.T) {
	assert.Equal(t, "a", findLongestCommonPrefix([]string{"aabc", "abc", "abd"}))
	assert.Equal(t, "ab", findLongestCommonPrefix([]string{"abbc", "abc", "abd"}))
	assert.Equal(t, "", findLongestCommonPrefix([]string{"abbc", "abc", "abd", "def"}))
}
