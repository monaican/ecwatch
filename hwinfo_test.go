package main

import "testing"

func TestBuildHWiNFOFanValues(t *testing.T) {
	t.Parallel()

	got := buildHWiNFOFanValues(fanSample{
		cpuRPM: 2516,
		gpuRPM: 75,
	})

	if len(got) != 2 {
		t.Fatalf("len(got)=%d, want 2", len(got))
	}

	if got[0].subKey != "Fan0" || got[0].name != "CPU Fan" || got[0].value != 2516 {
		t.Fatalf("unexpected Fan0 value: %+v", got[0])
	}

	if got[1].subKey != "Fan1" || got[1].name != "GPU Fan" || got[1].value != 75 {
		t.Fatalf("unexpected Fan1 value: %+v", got[1])
	}
}

func TestSanitizeHWiNFOGroupName(t *testing.T) {
	t.Parallel()

	got := sanitizeHWiNFOGroupName(`  EC/Watch\Test  `)
	if got != "EC_Watch_Test" {
		t.Fatalf("sanitizeHWiNFOGroupName()=%q, want %q", got, "EC_Watch_Test")
	}
}
