package main

import (
	"testing"

	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	t.Skip("not reliable on the CI")
	err := run([]string{"doctor", "."})
	if err != nil {
		t.Fatalf("err should be nil: %v", err)
	}

}

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}
