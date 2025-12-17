package assert

import (
	"reflect"
	"runtime"
)

// Nil panics if err is not nil.
func Nil(err error) {
	if err != nil {
		panic(err)
	}
}

// True panics if value is false.
func True(value bool, err error) {
	if !value {
		panic(err)
	}
}

// False panics if value is true.
func False(value bool, err error) {
	True(!value, err)
}

// NotNil asserts object is not nil, panic otherwise.
func NotNil(object interface{}) {
	True(!IsNil(object), nil)
}

// NotCircular detects circular dependency in singleton initialisation by
// inspecting current goroutine call stack; it panics if the same function
// appears twice in a row.
func NotCircular() {
	pc := make([]uintptr, 100)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])

	current, more := frames.Next()
	for more {
		next, m := frames.Next()
		more = m
		if current.Function == next.Function {
			panic("found circular dependency")
		}
	}
}

// IsNil reports whether object is nil, handling typed nil values.
func IsNil(object interface{}) bool {
	if object == nil {
		return true
	}
	value := reflect.ValueOf(object)
	switch value.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice, reflect.Func:
		return value.IsNil()
	}
	return false
}
