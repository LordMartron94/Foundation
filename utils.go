package foundation

import (
	"reflect"
	"runtime"
)

/*
	GetFunctionName returns the name of a function.

Pass the pointer to the function in.
*/
func GetFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}
