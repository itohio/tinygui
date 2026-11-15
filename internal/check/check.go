// Package check is the library's test assertions on plain `testing`. The
// tests used testify, whose reflection tricks panic under `tinygo test`
// ("reflect.Value has no flag field") before any test runs, so the
// packages' tests never ran on TinyGo (found in review). These
// helpers keep testify's call shapes (t first, want before got, optional
// message as a format string and its arguments) and stop the test on the
// first failure, as require does.
package check

import (
	"cmp"
	"fmt"
	"reflect"
	"testing"
)

// msgf formats testify-style trailing arguments: a format string and its
// arguments, or values printed in a row.
func msgf(msg []any) string {
	if len(msg) == 0 {
		return ""
	}
	if f, ok := msg[0].(string); ok && len(msg) > 1 {
		return ": " + fmt.Sprintf(f, msg[1:]...)
	}
	return ": " + fmt.Sprint(msg...)
}

// Equal fails unless want and got are deeply equal (same dynamic types).
func Equal(t testing.TB, want, got any, msg ...any) {
	t.Helper()
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("got %v (%T), want %v (%T)%s", got, got, want, want, msgf(msg))
	}
}

// True fails unless ok.
func True(t testing.TB, ok bool, msg ...any) {
	t.Helper()
	if !ok {
		t.Fatalf("false, want true%s", msgf(msg))
	}
}

// False fails if ok.
func False(t testing.TB, ok bool, msg ...any) {
	t.Helper()
	if ok {
		t.Fatalf("true, want false%s", msgf(msg))
	}
}

// isNil reports a nil interface, or an interface holding a nil pointer,
// slice, map, channel or function.
func isNil(v any) bool {
	if v == nil {
		return true
	}
	switch r := reflect.ValueOf(v); r.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer:
		return r.IsNil()
	}
	return false
}

// Nil fails unless v is nil (see isNil).
func Nil(t testing.TB, v any, msg ...any) {
	t.Helper()
	if !isNil(v) {
		t.Fatalf("got %v, want nil%s", v, msgf(msg))
	}
}

// NotNil fails if v is nil (see isNil).
func NotNil(t testing.TB, v any, msg ...any) {
	t.Helper()
	if isNil(v) {
		t.Fatalf("nil, want a value%s", msgf(msg))
	}
}

// Zero fails unless v is its type's zero value.
func Zero(t testing.TB, v any, msg ...any) {
	t.Helper()
	if v != nil && !reflect.ValueOf(v).IsZero() {
		t.Fatalf("got %v, want zero%s", v, msgf(msg))
	}
}

// Positive fails unless v > 0.
func Positive[T cmp.Ordered](t testing.TB, v T, msg ...any) {
	t.Helper()
	var zero T
	if v <= zero {
		t.Fatalf("got %v, want > 0%s", v, msgf(msg))
	}
}

// Greater fails unless a > b.
func Greater[T cmp.Ordered](t testing.TB, a, b T, msg ...any) {
	t.Helper()
	if a <= b {
		t.Fatalf("%v is not > %v%s", a, b, msgf(msg))
	}
}

// GreaterOrEqual fails unless a >= b.
func GreaterOrEqual[T cmp.Ordered](t testing.TB, a, b T, msg ...any) {
	t.Helper()
	if a < b {
		t.Fatalf("%v is not >= %v%s", a, b, msgf(msg))
	}
}

// LessOrEqual fails unless a <= b.
func LessOrEqual[T cmp.Ordered](t testing.TB, a, b T, msg ...any) {
	t.Helper()
	if a > b {
		t.Fatalf("%v is not <= %v%s", a, b, msgf(msg))
	}
}

// Len fails unless v (a slice, array, map, string or channel) has n
// elements.
func Len(t testing.TB, v any, n int, msg ...any) {
	t.Helper()
	if l := reflect.ValueOf(v).Len(); l != n {
		t.Fatalf("length %d, want %d%s", l, n, msgf(msg))
	}
}

// Empty fails unless v (a slice, array, map, string or channel) has no
// elements; nil counts as empty.
func Empty(t testing.TB, v any, msg ...any) {
	t.Helper()
	if v != nil && reflect.ValueOf(v).Len() != 0 {
		t.Fatalf("got %v, want empty%s", v, msgf(msg))
	}
}

// NotEmpty fails if v (a slice, array, map, string or channel) has no
// elements.
func NotEmpty(t testing.TB, v any, msg ...any) {
	t.Helper()
	if v == nil || reflect.ValueOf(v).Len() == 0 {
		t.Fatalf("empty, want elements%s", msgf(msg))
	}
}

// Contains fails unless s holds v.
func Contains[T comparable](t testing.TB, s []T, v T, msg ...any) {
	t.Helper()
	for _, e := range s {
		if e == v {
			return
		}
	}
	t.Fatalf("%v doesn't contain %v%s", s, v, msgf(msg))
}

// Panics fails unless f panics.
func Panics(t testing.TB, f func(), msg ...any) {
	t.Helper()
	if _, ok := recovered(f); !ok {
		t.Fatalf("no panic, want one%s", msgf(msg))
	}
}

// PanicsWithValue fails unless f panics with a value deeply equal to want.
func PanicsWithValue(t testing.TB, want any, f func(), msg ...any) {
	t.Helper()
	got, ok := recovered(f)
	if !ok {
		t.Fatalf("no panic, want %v%s", want, msgf(msg))
	}
	if !reflect.DeepEqual(want, got) {
		t.Fatalf("panicked with %v, want %v%s", got, want, msgf(msg))
	}
}

// recovered runs f and returns what it panicked with, if it did.
func recovered(f func()) (v any, panicked bool) {
	defer func() {
		if r := recover(); r != nil || panicked {
			v, panicked = r, true
		}
	}()
	panicked = true
	f()
	panicked = false
	return nil, false
}
