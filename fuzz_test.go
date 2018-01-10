// +build gofuzz

package lazylist

import (
	"testing"
)

func TestFuzz(t *testing.T) {
	Fuzz([]byte{6, 48, 109})
}
