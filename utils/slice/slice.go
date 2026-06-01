package slice

import (
	"fmt"
	"iter"
	"slices"
	"strings"
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

func Joins(slice []string, with string) string {
	builder := strings.Builder{}

	if len(slice) == 0 {
		return ""
	}

	builder.WriteString(slice[0])

	for _, item := range slice[1:] {
		builder.WriteString(with)
		builder.WriteString(item)
	}

	return builder.String()
}

func Joinf[T fmt.Stringer](slice []T, with string) string {
	builder := strings.Builder{}

	if len(slice) == 0 {
		return ""
	}

	builder.WriteString(slice[0].String())

	for _, item := range slice[1:] {
		builder.WriteString(with)
		builder.WriteString(item.String())
	}

	return builder.String()
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

func TryMap[T any, U any](slice []T, f func(T) (U, error)) ([]U, error) {
	result := []U{}

	for _, t := range slice {
		u, err := f(t)

		if err != nil {
			return nil, err
		}

		result = append(result, u)
	}

	return result, nil
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

func Ref[T any](slice []T) []*T {
	result := make([]*T, len(slice))
	for i := range slice {
		result[i] = &slice[i]
	}

	return result
}
