package itertools

import "iter"

func Collect[T any](i iter.Seq[T]) []T {
	var result []T

	next, stop := iter.Pull(i)
	defer stop()

	for v, ok := next(); ok; v, ok = next() {
		result = append(result, v)
	}

	return result
}
