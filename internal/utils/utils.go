package utils

import (
	"time"

	"github.com/pkg/errors"
)

// CurrentTimestamp is a utility method to make sure UTC time is used all over the code
func CurrentTimestamp() time.Time {
	return time.Now().UTC()
}

func RunEvery(d time.Duration, fs ...func() error) error {
	if err := runFuncs(CurrentTimestamp(), fs...); err != nil {
		return errors.Wrap(err, "failed to run funcs initial")
	}
	for x := range time.Tick(d) {
		if err := runFuncs(x, fs...); err != nil {
			return errors.Wrap(err, "failed to run funcs")
		}
	}
	return nil
}

func runFuncs(x time.Time, fs ...func() error) error {
	for i, f := range fs {
		if err := f(); err != nil {
			return errors.Wrapf(err, "failed to run function: %v -> %v", x, i)
		}
	}
	return nil
}

func Map[T any](collection []T, fn func(T) T) []T {
	res := make([]T, len(collection))
	for i, el := range collection {
		res[i] = fn(el)
	}
	return res
}

func ToCommand(s string) string {
	return "/" + s
}
