package itertools

import "iter"

func ForEach[T any](i iter.Seq[T], fn func(T)) {
	for t := range i {
		fn(t)
	}
}
