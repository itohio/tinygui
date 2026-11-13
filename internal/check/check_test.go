package check

import (
	"errors"
	"testing"
)

// tb records a failure instead of stopping the test.
type tb struct {
	testing.TB
	failed bool
}

func (r *tb) Helper()               {}
func (r *tb) Fatalf(string, ...any) { r.failed = true }

// Each helper passes on a right value and fails on a wrong one: a helper
// that never failed would let every test that uses it pass.
func TestHelpersFailOnWrongValues(t *testing.T) {
	type box struct{ v []int }
	var nilPtr *box
	var nilSlice []int
	cases := []struct {
		name      string
		pass, bad func(t testing.TB)
	}{
		{"Equal", func(t testing.TB) { Equal(t, box{[]int{1}}, box{[]int{1}}) }, func(t testing.TB) { Equal(t, int16(3), 3) }},
		{"True", func(t testing.TB) { True(t, true) }, func(t testing.TB) { True(t, false) }},
		{"False", func(t testing.TB) { False(t, false) }, func(t testing.TB) { False(t, true) }},
		{"Nil", func(t testing.TB) { Nil(t, nilPtr) }, func(t testing.TB) { Nil(t, &box{}) }},
		{"NotNil", func(t testing.TB) { NotNil(t, errors.New("x")) }, func(t testing.TB) { NotNil(t, nilSlice) }},
		{"Zero", func(t testing.TB) { Zero(t, 0.0) }, func(t testing.TB) { Zero(t, 0.5) }},
		{"Positive", func(t testing.TB) { Positive(t, 1) }, func(t testing.TB) { Positive(t, 0) }},
		{"Greater", func(t testing.TB) { Greater(t, 2, 1) }, func(t testing.TB) { Greater(t, 1, 1) }},
		{"GreaterOrEqual", func(t testing.TB) { GreaterOrEqual(t, 1, 1) }, func(t testing.TB) { GreaterOrEqual(t, 0, 1) }},
		{"LessOrEqual", func(t testing.TB) { LessOrEqual(t, 1, 1) }, func(t testing.TB) { LessOrEqual(t, 2, 1) }},
		{"Len", func(t testing.TB) { Len(t, []int{1, 2}, 2) }, func(t testing.TB) { Len(t, "abc", 2) }},
		{"Empty", func(t testing.TB) { Empty(t, nilSlice) }, func(t testing.TB) { Empty(t, map[int]int{1: 1}) }},
		{"NotEmpty", func(t testing.TB) { NotEmpty(t, "a") }, func(t testing.TB) { NotEmpty(t, []int{}) }},
		{"Contains", func(t testing.TB) { Contains(t, []string{"a", "b"}, "b") }, func(t testing.TB) { Contains(t, []string{"a"}, "b") }},
		{"Panics", func(t testing.TB) { Panics(t, func() { panic("x") }) }, func(t testing.TB) { Panics(t, func() {}) }},
		{"PanicsWithValue", func(t testing.TB) { PanicsWithValue(t, "x", func() { panic("x") }) }, func(t testing.TB) { PanicsWithValue(t, "x", func() { panic("y") }) }},
	}
	for _, c := range cases {
		ok, bad := &tb{TB: t}, &tb{TB: t}
		c.pass(ok)
		c.bad(bad)
		if ok.failed {
			t.Errorf("%s failed on a right value", c.name)
		}
		if !bad.failed {
			t.Errorf("%s passed a wrong value", c.name)
		}
	}
}
