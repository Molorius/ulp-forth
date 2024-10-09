package forth

import (
	_ "embed"
	"fmt"
	"testing"
)

//go:embed test/suite_test.f
var suite string

func TestSuite(t *testing.T) {
	tests := []struct {
		name string // the name of the test
		code string // any code we want to run in the "main" word
		// tests that run in the global context should be added to suite_test.f
	}{
		// ABORT
		// ABORT"
		// ABS
		// ACCEPT
		// ACTION-OF
		// AGAIN
		// ALIGN
		// ALIGNED
		// ALLOT
		// AND
		// BASE
		// BEGIN doesn't have tests
		{
			name: "BL",
			code: "T{ BL -> 32 }T", // TODO fix after hex printing is enabled
		},
		{
			name: "[",
			code: "T{ GC3 -> 88 }T", // TODO fix after hex printing is enabled
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := createTest(tt.code)
			runOutputTest(code, "", t)
		})
	}
}

func createTest(code string) string {
	return fmt.Sprintf("%s : MAIN %s ESP.DONE ; ", suite, code)
}
