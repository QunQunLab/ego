package utils

import "testing"

func TestTryCatch(t *testing.T) {
	type MyError struct {
		error
	}

	TryCatch{}.Try(func() {
		t.Log("do something buggy")
		panic(MyError{})
	}).Catch(MyError{}, func(err error) {
		t.Log("catch MyError")
	}).CatchAll(func(err error) {
		t.Log("catch error")
	}).Finally(func() {
		t.Log("finally do something")
	})
	t.Log("done")
}
