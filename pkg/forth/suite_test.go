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
		{
			name: "BASE",
			code: "T{ GN2 -> 10 A }T",
		},
		// BEGIN doesn't have tests
		{
			name: "BL",
			code: "T{ BL -> 20 }T",
		},
		// BUFFER
		{
			name: "[",
			code: "T{ GC3 -> 58 }T",
		},
		// [CHAR]
		// [COMPILE]
		{
			name: "[']",
			code: "T{ POSTPONE GT2 EXECUTE -> 123 }T", // postpone because it's inside "main"
		},
		// CASE
		// C,
		// CELL+ doesn't have regular tests
		// CELLS
		// C@
		// CHAR
		// CHAR+
		// CHARS
		// COMPILE,
		{
			name: "CONSTANT",
			code: `
				T{ X123 -> 123 }T
				T{ Y123 -> 123 }T
			`,
		},
		// COUNT
		// CR does not have any tests
		// CREATE does not have any tests
		// C! does not have any tests
		{
			name: ":",
			code: `
				T{ NOP1 -> }T
				T{ NOP2 -> }T
				T{ GDX -> 123 234 }T
			`,
		},
		{
			name: ":NONAME",
			code: `
				T{ nn1 @ EXECUTE -> 1234 }T
				T{ nn2 @ EXECUTE -> 9876 }T
			`,
		},
		// ,
		// C"
		// DECIMAL does not have any tests
		// DEFER
		// DEFER@
		// DEFER!
		{
			name: "DEPTH",
			code: `
				T{ 0 1 DEPTH -> 0 1 2 }T
				T{   0 DEPTH -> 0 1   }T
				T{     DEPTH -> 0     }T
			`,
		},
		// DO does not have any tests
		// DOES>
		{
			name: "DROP",
			code: `
				T{ 1 2 DROP -> 1 }T
				T{ 0   DROP ->   }T
			`,
		},
		{
			name: "DUP",
			code: `
				T{ 1 DUP -> 1 1 }T
			`,
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
