package logger

import (
	"bytes"
	"testing"
)

func TestWriteStartupBanner(t *testing.T) {
	var output bytes.Buffer

	err := WriteStartupBanner(
		&output,
		"0.1.0",
		"5074874",
		"2026-09-02T09:47:06Z",
	)
	if err != nil {
		t.Fatalf("write startup banner: %v", err)
	}

	expected := "" +
		"============================================================\n" +
		"                          EXPONIA\n" +
		"============================================================\n" +
		"version:    0.1.0\n" +
		"commit:     5074874\n" +
		"build_time: 2026-09-02T09:47:06Z\n" +
		"============================================================\n"

	if output.String() != expected {
		t.Fatalf("unexpected startup banner:\n%s", output.String())
	}
}
