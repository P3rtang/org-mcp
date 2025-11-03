package result

type Result[T any] struct {
	value T
	err   error
	ok    bool
}

func TryFunc[T any](fn func() (T, error)) Result[T] {
	value, err := fn()

	if err != nil {
		return Err[T](err)
	}

	return Ok(value)
}

func Err[T any](err error) Result[T] {
	return Result[T]{err: err}
}

func Ok[T any](t T) Result[T] {
	return Result[T]{
		value: t,
		ok:    true,
	}
}

/*
Returns true if the Option does not contain a value, false otherwise.

Inverse of IsSome
*/
func (o *Result[T]) IsErr() bool {
	return !o.ok
}

/*
Returns true if the Option contains a value, false otherwise.

Inverse of IsNone
*/
func (o *Result[T]) IsOk() bool {
	return o.ok
}

func (o Result[T]) Split() (T, error) {
	return o.value, o.err
}

/*
This function returns a copy of the data and not a pointer to it.
Panics if the Option is None.

Modifying the result will not modify the original data
instead use UnwrapPtr or Map
*/
func (o Result[T]) Unwrap() T {
	if o.IsErr() {
		panic("called Unwrap on a None value")
	}

	return o.value
}

func (o Result[T]) UnwrapOr(t T) T {
	if o.IsErr() {
		return t
	}

	return o.value
}

/*
This function returns a pointer to the data contained in the Option.
Panics if the Option is None.
*/
func (o Result[T]) UnwrapPtr() *T {
	if o.IsErr() {
		panic("called Unwrap on a None value")
	}

	return &o.value
}

/*
 */
func (o Result[T]) Then(fn func(T)) {
	if o.IsErr() {
		return
	}

	fn(o.value)
}

func (o Result[T]) AndThen(fn func(T) bool) bool {
	if o.IsErr() {
		return false
	}

	return fn(o.value)
}

/*
This function applies the given function to the Option value and returns the result.
*/
func Map[T any, U any](r Result[T], fn func(T) U) Result[U] {
	if r.IsErr() {
		return Result[U]{err: r.err}
	}

	return Ok(fn(r.value))
}

func Try[T any, U any](r Result[T], fn func(T) Result[U]) Result[U] {
	if r.IsErr() {
		return Result[U]{err: r.err}
	}

	return fn(r.value)
}
