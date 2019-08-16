package a

import (
	"fmt"
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
	if !cmp.Equal(got, want) { // error or something
		fmt.Println("but they're not equal!")
	}
}
