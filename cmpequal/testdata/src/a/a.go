package a

import (
	"fmt"
	"testing"

 "github.com/google/go-cmp/cmp"
)

func TestSomething(t *testing.T) {
	var x int = 0
	var y int = 0
	if ! cmp.Equal(x, &y) { // error or something
		fmt.Println("but they're not equal!")
	}
}