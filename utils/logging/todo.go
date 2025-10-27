package logging

import (
	"fmt"
	"runtime"
)

func TODO[T any]() T {
	_, file, line, _ := runtime.Caller(1)
	println(fmt.Sprintf("%s:%d: TODO: not implemented yet", file, line))
	panic("")
}
