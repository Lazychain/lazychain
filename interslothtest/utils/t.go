package utils

import (
	"fmt"
	"time"
)

var _ CustomT = &FakeT{}

type CustomT interface {
	Name() string
	Cleanup(func())

	Skip(...any)

	Parallel()

	Log(args ...any)
	Logf(format string, args ...any)
	Errorf(string, ...interface{})

	Fail()
	FailNow()

	Helper()

	Failed() bool
	Skipped() bool
}

type FakeT struct {
	FakeName     string
	CleanupFuncs []func()
}

func (f *FakeT) Errorf(s string, i ...interface{}) {
	fmt.Printf(s, i...)
}

func (f *FakeT) Fail() {
	fmt.Println("Failed!")
}

func (f *FakeT) FailNow() {
	panic("FailNow!")
}

func (f *FakeT) Skip(a ...any) {}

func (f *FakeT) Parallel() {}

func (f *FakeT) Skipped() bool {
	return false
}

func (f *FakeT) Name() string {
	return f.FakeName
}

func (f *FakeT) Helper() {}

func (f *FakeT) Failed() bool {
	return false
}

func (f *FakeT) Cleanup(foo func()) {
	f.CleanupFuncs = append(f.CleanupFuncs, foo)
}

func (f *FakeT) ActuallyRunCleanups() {
	fmt.Println("Actually running cleanups from FakeT...")
	for _, foo := range f.CleanupFuncs {
		foo()
	}

	time.Sleep(1 * time.Second)

	f.CleanupFuncs = []func(){}
}

func (f *FakeT) Log(args ...any) {
	fmt.Println(args...)
}

func (f *FakeT) Logf(format string, args ...any) {
	fmt.Printf(format, args...)
}
