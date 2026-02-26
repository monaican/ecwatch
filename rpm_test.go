package main

import "testing"

func TestRPMFromBytes(t *testing.T) {
	t.Parallel()

	got := rpmFromBytes(0x0A, 0xF0)
	if got != 2800 {
		t.Fatalf("rpmFromBytes(0x0A,0xF0)=%d, want 2800", got)
	}
}
