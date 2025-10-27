package itertools

import (
	"iter"
)

func Filter[T any](seq iter.Seq[T], f func(T) bool) iter.Seq[T] {
	next, stop := iter.Pull(seq)

	return func(yield func(T) bool) {
		defer stop()

		for t, ok := next(); ok; t, ok = next() {
			if f(t) {
				if !yield(t) {
					return
				}
			}
		}
	}
}
