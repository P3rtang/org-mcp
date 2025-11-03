package option

import "reflect"

type Option[T any] struct {
	value T
	set   bool
}

func None[T any]() Option[T] {
	return Option[T]{}
}

func Some[T any](t T) Option[T] {
	return Option[T]{
		value: t,
		set:   true,
	}
}

func From[T *any](t T) Option[T] {
	if t == nil {
		return None[T]()
	}

	return Some(t)
}

/*
Returns true if the Option does not contain a value, false otherwise.

Inverse of IsSome
*/
func (o *Option[T]) IsNone() bool {
	return !o.set
}

/*
Returns true if the Option contains a value, false otherwise.

Inverse of IsNone
*/
func (o *Option[T]) IsSome() bool {
	return o.set
}

func (o Option[T]) Split() (T, bool) {
	return o.value, o.set
}

/*
This function returns a copy of the data and not a pointer to it.
Panics if the Option is None.

Modifying the result will not modify the original data
instead use UnwrapPtr or Map
*/
func (o Option[T]) Unwrap() T {
	if o.IsNone() {
		panic("called Unwrap on a None value")
	}

	return o.value
}

/*
This function returns a pointer to the data contained in the Option.
Panics if the Option is None.
*/
func (o Option[T]) UnwrapPtr() *T {
	if o.IsNone() {
		panic("called Unwrap on a None value")
	}

	return &o.value
}

func (o Option[T]) UnwrapOr(t T) T {
	if o.IsNone() {
		return t
	}

	return o.value
}

func (o Option[T]) UnwrapOrElse(fn func() T) T {
	if o.IsNone() {
		return fn()
	}

	return o.value
}

/*
 */
func (o Option[T]) Then(fn func(T)) {
	if o.IsNone() {
		return
	}

	fn(o.value)
}

func (o Option[T]) AndThen(fn func(T) bool) bool {
	if o.IsNone() {
		return false
	}

	return fn(o.value)
}

/*
This function applies the given function to the Option value and returns the result.
*/
func Map[T any, U any](option Option[T], fn func(T) U) Option[U] {
	if option.IsNone() {
		return Option[U]{}
	}

	return Some(fn(option.value))
}

func Cast[T any, U any](o Option[T]) Option[U] {
	if o.IsNone() {
		return Option[U]{}
	}

	if reflect.TypeOf(o.value) == reflect.TypeOf((*U)(nil)).Elem() {
		return Some(any(o.value).(U))
	}

	return Option[U]{}
}
