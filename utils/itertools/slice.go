package itertools

import "iter"

func FromSlice[T any](slice []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, t := range slice {
			if !yield(t) {
				return
			}
		}
	}
}
