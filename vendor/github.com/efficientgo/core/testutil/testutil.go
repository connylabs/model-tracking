// Copyright (c) The EfficientGo Authors.
// Licensed under the Apache License 2.0.

package testutil

import (
	"fmt"
	"path/filepath"
	"reflect"
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/efficientgo/core/errors"
	"github.com/efficientgo/core/testutil/internal"
)

// Assert fails the test if the condition is false.
func Assert(tb testing.TB, condition bool, v ...interface{}) {
	tb.Helper()
	if condition {
		return
	}
	_, file, line, _ := runtime.Caller(1)

	var msg string
	if len(v) > 0 {
		msg = fmt.Sprintf(v[0].(string), v[1:]...)
	}
	tb.Fatalf("\033[31m%s:%d: "+msg+"\033[39m\n\n", filepath.Base(file), line)
}

// Ok fails the test if an err is not nil.
func Ok(tb testing.TB, err error, v ...interface{}) {
	tb.Helper()
	if err == nil {
		return
	}
	_, file, line, _ := runtime.Caller(1)

	var msg string
	if len(v) > 0 {
		msg = fmt.Sprintf(v[0].(string), v[1:]...)
	}
	tb.Fatalf("\033[31m%s:%d:"+msg+"\n\n unexpected error: %s\033[39m\n\n", filepath.Base(file), line, err.Error())
}

// NotOk fails the test if an err is nil.
func NotOk(tb testing.TB, err error, v ...interface{}) {
	tb.Helper()
	if err != nil {
		return
	}
	_, file, line, _ := runtime.Caller(1)

	var msg string
	if len(v) > 0 {
		msg = fmt.Sprintf(v[0].(string), v[1:]...)
	}
	tb.Fatalf("\033[31m%s:%d:"+msg+"\n\n expected error, got nothing \033[39m\n\n", filepath.Base(file), line)
}

// Equals fails the test if exp is not equal to act.
func Equals(tb testing.TB, exp, act interface{}, v ...interface{}) {
	tb.Helper()
	if reflect.DeepEqual(exp, act) {
		return
	}
	_, file, line, _ := runtime.Caller(1)

	var msg string
	if len(v) > 0 {
		msg = fmt.Sprintf(v[0].(string), v[1:]...)
	}
	tb.Fatal(sprintfWithLimit("\033[31m%s:%d:"+msg+"\n\n\texp: %#v\n\n\tgot: %#v%s\033[39m\n\n", filepath.Base(file), line, exp, act, diff(exp, act)))
}

// FaultOrPanicToErr returns error if panic of fault was triggered during execution of function.
func FaultOrPanicToErr(f func()) (err error) {
	// Set this go routine to panic on segfault to allow asserting on those.
	debug.SetPanicOnFault(true)
	defer func() {
		if r := recover(); r != nil {
			err = errors.Newf("invoked function panicked or caused segmentation fault: %v", r)
		}
		debug.SetPanicOnFault(false)
	}()

	f()

	return err
}

func sprintfWithLimit(act string, v ...interface{}) string {
	s := fmt.Sprintf(act, v...)
	if len(s) > 10000 {
		return s[:10000] + "...(output trimmed)"
	}
	return s
}

func typeAndKind(v interface{}) (reflect.Type, reflect.Kind) {
	t := reflect.TypeOf(v)
	k := t.Kind()

	if k == reflect.Ptr {
		t = t.Elem()
		k = t.Kind()
	}
	return t, k
}

// diff returns a diff of both values as long as both are of the same type and
// are a struct, map, slice, array or string. Otherwise it returns an empty string.
func diff(expected interface{}, actual interface{}) string {
	if expected == nil || actual == nil {
		return ""
	}

	et, ek := typeAndKind(expected)
	at, _ := typeAndKind(actual)
	if et != at {
		return ""
	}

	if ek != reflect.Struct && ek != reflect.Map && ek != reflect.Slice && ek != reflect.Array && ek != reflect.String {
		return ""
	}

	var e, a string
	c := spew.ConfigState{
		Indent:                  " ",
		DisablePointerAddresses: true,
		DisableCapacities:       true,
		SortKeys:                true,
	}
	if et != reflect.TypeOf("") {
		e = c.Sdump(expected)
		a = c.Sdump(actual)
	} else {
		e = reflect.ValueOf(expected).String()
		a = reflect.ValueOf(actual).String()
	}

	diff, _ := internal.GetUnifiedDiffString(internal.UnifiedDiff{
		A:        internal.SplitLines(e),
		B:        internal.SplitLines(a),
		FromFile: "Expected",
		FromDate: "",
		ToFile:   "Actual",
		ToDate:   "",
		Context:  1,
	})
	return "\n\nDiff:\n" + diff
}
