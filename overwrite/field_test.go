package overwrite_test

import (
	"testing"

	"github.com/morikuni/go-experiment/overwrite"
	"github.com/morikuni/go-experiment/overwrite/internal"
)

func TestField(t *testing.T) {
	x := internal.NewX("aaa")
	if got, want := x.Get(), "aaa"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	overwrite.Field(&x, "val", "bbb")
	if got, want := x.Get(), "bbb"; got != want {
		t.Errorf("got %v, want %v", got, want)
	}

	err := overwrite.Field(&x, "foo", "bbb")
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
