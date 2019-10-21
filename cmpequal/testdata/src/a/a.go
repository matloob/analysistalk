// Package a is a good package.
package a

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

type X struct {
}

func NewX() *X {
	return &X{}
}

func TestSomething(t *testing.T) {
	want := X{}
	got := NewX()
	if !cmp.Equal(got, want) {
		t.Error("but they're not equal!")
	}
}
