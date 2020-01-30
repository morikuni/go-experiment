package overwrite_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/morikuni/go-experiment/overwrite"
	"github.com/morikuni/go-experiment/overwrite/internal"
)

func TestField(t *testing.T) {
	errEEE := errors.New("eee")
	x := internal.NewX("aaa", errEEE)
	if got, want := x.GetVal(), "aaa"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := x.GetErr(), errEEE; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	err := overwrite.Field(&x, "val", "bbb")
	if got, want := err, error(nil); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := x.GetVal(), "bbb"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	errFFF := errors.New("fff")
	err = overwrite.Field(&x, "err", errFFF)
	if got, want := err, error(nil); got != want {
		t.Errorf("got %v, want %v", got, want)
	}
	if got, want := x.GetErr(), errFFF; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	err = overwrite.Field(&x, "foo", "bbb")
	if got, want := err, overwrite.ErrNoSuchField; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	err = overwrite.Field(&x, "val", 123)
	if got, want := err, overwrite.ErrTypeMismatch; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	var i int
	err = overwrite.Field(&i, "val", 123)
	if got, want := err, overwrite.ErrNotStruct; got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

type Foo struct {
	err error
}

func TestFoo(t *testing.T) {
	sErr := Foo{}
	err := fmt.Errorf("this is error")

	e := overwrite.Field(&sErr, "err", err)
	if e != nil {
		t.Fatal(e)
	}
	if err != sErr.err {
		t.Fatal(e)
	}
}
