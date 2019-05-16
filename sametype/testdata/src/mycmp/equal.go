package mycmp

import (
	"annotate"
	"reflect"
)

func Equal(a, b interface{}) bool {
	annotate.SameType()
	return reflect.DeepEqual(a, b)
}