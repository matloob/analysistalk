package a

import (
	"fmt"
	"testing"

  "mycmp"
)

func TestSomething(t *testing.T) {
	var x int = 0
	var y int = 0
	if !mycmp.Equal(x, &y) { // want "Calls to Equal must have arguments of the same type"
		fmt.Println("but they're not equal!")
	}
}