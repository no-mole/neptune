package errgroup

import (
	"errors"
	"testing"
)

func TestGroup_Wait(t *testing.T) {
	errMsg := "error panic"
	eg := Group{}
	eg.Go(func() error {
		return nil
	})
	eg.Go(func() error {
		panic(errMsg)
		return nil
	})
	err := eg.Wait()
	if err == nil {
		t.Fatal("didn't catch err")
	}
	if err.Error() != errMsg {
		t.Fatal("didn't catch err")
	}
}
func TestGroup_Wait1(t *testing.T) {
	expectedErr := errors.New("xxx")
	errMsg := "error panic"
	ch := make(chan struct{})
	eg := Group{}
	eg.Go(func() error {
		ch <- struct{}{}
		return expectedErr
	})
	eg.Go(func() error {
		<-ch
		panic(errMsg)
		return nil
	})
	err := eg.Wait()
	if err != expectedErr {
		t.Fatal("didn't catch expectedErr")
	}
	if err.Error() != expectedErr.Error() {
		t.Fatal("didn't catch err")
	}
}
