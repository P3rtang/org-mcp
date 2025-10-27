package itertools

import "iter"

func Map[T any, U any](i iter.Seq[T], mapping func(T) U) iter.Seq[U] {
	next, stop := iter.Pull(i)

	return func(yield func(U) bool) {
		defer stop()

		for t, ok := next(); ok; t, ok = next() {
			if !yield(mapping(t)) {
				return
			}
		}
	}
}
