package itertools

import (
	"github.com/p3rtang/org-mcp/utils/option"
	"iter"
)

func Find[T any](seq iter.Seq[T], f func(T) bool) option.Option[T] {
	for t := range seq {
		if f(t) {
			return option.Some(t)
		}
	}

	return option.None[T]()
}
