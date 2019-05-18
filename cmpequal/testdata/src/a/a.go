package a

import (
	"fmt"
	"testing"

 "github.com/google/go-cmp/cmp"
)

type X struct {
}

func TestSomething(t *testing.T) {
	var x X
	var y X
	if ! cmp.Equal(x, &y) { // error or something
		fmt.Println("but they're not equal!")
	}
}
