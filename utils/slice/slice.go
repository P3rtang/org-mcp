package slice

import (
	"iter"
	"slices"
)

func Any[T any](slice []T, fn func(T) bool) bool {
	return slices.ContainsFunc(slice, fn)
}

func All[T any](slice []T, fn func(T) bool) bool {
	for _, item := range slice {
		if !fn(item) {
			return false
		}
	}

	return true
}

func Chars(str string) iter.Seq[rune] {
	return func(yield func(rune) bool) {
		for _, char := range str {
			if !yield(char) {
				return
			}
		}
	}
}

func Map[T any, U any](slice []T, f func(T) U) []U {
	result := []U{}

	for _, t := range slice {
		result = append(result, f(t))
	}

	return result
}

func Reduce[T any, U any](slice []T, f func(U, T) U, acc U) U {
	for _, t := range slice {
		acc = f(acc, t)
	}

	return acc
}

func Filter[T any](slice []T, f func(T) bool) []T {
	return Reduce(slice, func(acc []T, t T) []T {
		if f(t) {
			acc = append(acc, t)
		}

		return acc
	}, []T{})
}
