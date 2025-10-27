package itertools

import (
	"iter"
	"slices"
)

func Dedup[T any, U comparable](seq iter.Seq[T], key func(T) U) iter.Seq[T] {
	keys := []U{}
	next, stop := iter.Pull(seq)

	return func(yield func(T) bool) {
		defer stop()

		for t, ok := next(); ok; t, ok = next() {
			k := key(t)

			if slices.Contains(keys, k) {
				continue
			}

			keys = append(keys, k)

			if !yield(t) {
				return
			}
		}
	}
}
