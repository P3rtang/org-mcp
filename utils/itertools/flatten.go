package itertools

import "iter"

func Flatten[T any](seq iter.Seq[[]T]) iter.Seq[T] {
	next, stop := iter.Pull(seq)

	return func(yield func(T) bool) {
		for slice, ok := next(); ok; slice, ok = next() {
			for _, t := range slice {
				if !yield(t) {
					stop()
					return
				}
			}
		}
	}
}
