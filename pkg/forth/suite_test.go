package forth

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/Molorius/ulp-c/pkg/asm"
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
		{
			name: "ABS",
			code: `
				T{       0 ABS ->          0 }T
				T{       1 ABS ->          1 }T
				T{      -1 ABS ->          1 }T
				T{ MIN-INT ABS -> MID-UINT+1 }T
			`,
		},
		// ACCEPT
		{
			name: "ACTION-OF",
			code: `
				\ uses ['] rather than ' so we can run on ulp
				T{ ['] * ['] defer1 DEFER! ->   }T
				T{          2 3 defer1 -> 6 }T
				T{ ACTION-OF defer1 -> ['] * }T
				T{    action-defer1 -> ['] * }T

				T{ ['] + IS defer1 ->   }T
				T{    1 2 defer1 -> 3 }T
				T{ ACTION-OF defer1 -> ['] + }T
				T{    action-defer1 -> ['] + }T
			`,
		},
		// AGAIN
		// ALIGN
		// ALIGNED
		// ALLOT
		{
			name: "AND",
			code: `
				T{ 0 0 AND -> 0 }T
				T{ 0 1 AND -> 0 }T
				T{ 1 0 AND -> 0 }T
				T{ 1 1 AND -> 1 }T
				T{ 0 INVERT 1 AND -> 1 }T
				T{ 1 INVERT 1 AND -> 0 }T

				T{ 0S 0S AND -> 0S }T
				T{ 0S 1S AND -> 0S }T
				T{ 1S 0S AND -> 0S }T
				T{ 1S 1S AND -> 1S }T
			`,
		},
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
		{
			name: "[COMPILE]",
			code: `
				T{ 123 [COMPILE] [c1] -> 123 123 }T
				T{ 234 [c2] -> 234 234 }T
				T{ -1 [c3] -> 111 }T
				T{  0 [c3] -> 222 }T
			`,
		},
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
		{
			name: "COUNT",
			code: `
				T{ GT1STRING COUNT -> GT1STRING CHAR+ 3 }T
			`,
		},
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
		{
			name: "DEFER",
			code: `
				\ uses ['] rather than ' so we can run on ulp
				T{ ['] * ['] defer2 DEFER! -> }T
				T{   2 3 defer2 -> 6    }T
				T{ ['] + IS defer2 ->   }T
				T{    1 2 defer2 -> 3   }T
			`,
		},
		{
			name: "DEFER@",
			code: `
				\ uses ['] rather than ' so we can run on ulp
				T{ ['] * ['] defer4 DEFER! -> }T
				T{ 2 3 defer4 -> 6 }T
				T{ ['] defer4 DEFER@ -> ['] * }T

				T{ ['] + IS defer4 -> }T
				T{ 1 2 defer4 -> 3 }T
				T{ ['] defer4 DEFER@ -> ['] + }T
			`,
		},
		{
			name: "DEFER!",
			code: `
				\ uses ['] rather than ' so we can run on ulp
				T{ ['] * ['] defer3 DEFER! -> }T
				T{ 2 3 defer3 -> 6 }T

				T{ ['] + ['] defer3 DEFER! -> }T
				T{ 1 2 defer3 -> 3 }T
			`,
		},
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
		{
			name: "OR",
			code: `
				T{ 0S 0S OR -> 0S }T
				T{ 0S 1S OR -> 1S }T
				T{ 1S 0S OR -> 1S }T
				T{ 1S 1S OR -> 1S }T
			`,
		},
		{
			name: "*",
			code: `
				T{  0  0 * ->  0 }T \ TEST IDENTITIES
				T{  0  1 * ->  0 }T
				T{  1  0 * ->  0 }T
				T{  1  2 * ->  2 }T
				T{  2  1 * ->  2 }T
				T{  3  3 * ->  9 }T
				T{ -3  3 * -> -9 }T
				T{  3 -3 * -> -9 }T
				T{ -3 -3 * ->  9 }T
				T{ MID-UINT+1 1 RSHIFT 2 *               -> MID-UINT+1 }T
				T{ MID-UINT+1 2 RSHIFT 4 *               -> MID-UINT+1 }T
				T{ MID-UINT+1 1 RSHIFT MID-UINT+1 OR 2 * -> MID-UINT+1 }T
			`,
		},
		{
			name: "S\"",
			code: `
				T{ GC4 SWAP DROP  -> 2 }T
				T{ GC4 DROP DUP C@ SWAP CHAR+ C@ -> 58 59 }T
				T{ GC5 -> }T
			`,
		},
		{
			name: "IS",
			code: `
				\ uses ['] rather than ' so we can run on ulp
				T{ ['] * IS defer5 -> }T
				T{ 2 3 defer5 -> 6    }T

				T{ ['] + is-defer5 -> }T
				T{ 1 2 defer5 -> 3    }T
			`,
		},
		{
			name: "U<",
			code: `
				T{        0        1 U< -> <TRUE>  }T
				T{        1        2 U< -> <TRUE>  }T
				T{        0 MID-UINT U< -> <TRUE>  }T
				T{        0 MAX-UINT U< -> <TRUE>  }T
				T{ MID-UINT MAX-UINT U< -> <TRUE>  }T
				T{        0        0 U< -> <FALSE> }T
				T{        1        1 U< -> <FALSE> }T
				T{        1        0 U< -> <FALSE> }T
				T{        2        1 U< -> <FALSE> }T
				T{ MID-UINT        0 U< -> <FALSE> }T
				T{ MAX-UINT        0 U< -> <FALSE> }T
				T{ MAX-UINT MID-UINT U< -> <FALSE> }T
			`,
		},
		{
			name: "U>",
			code: `
				T{        0        1 U> -> <FALSE> }T
				T{        1        2 U> -> <FALSE> }T
				T{        0 MID-UINT U> -> <FALSE> }T
				T{        0 MAX-UINT U> -> <FALSE> }T
				T{ MID-UINT MAX-UINT U> -> <FALSE> }T
				T{        0        0 U> -> <FALSE> }T
				T{        1        1 U> -> <FALSE> }T
				T{        1        0 U> -> <TRUE>  }T
				T{        2        1 U> -> <TRUE>  }T
				T{ MID-UINT        0 U> -> <TRUE>  }T
				T{ MAX-UINT        0 U> -> <TRUE>  }T
				T{ MAX-UINT MID-UINT U> -> <TRUE>  }T
			`,
		},
		{
			name: "XOR",
			code: `
				T{ 0S 0S XOR -> 0S }T
				T{ 0S 1S XOR -> 1S }T
				T{ 1S 0S XOR -> 1S }T
				T{ 1S 1S XOR -> 0S }T
			`,
		},
		{
			name: "0=",
			code: `
				T{        0 0= -> <TRUE>  }T
				T{        1 0= -> <FALSE> }T
				T{        2 0= -> <FALSE> }T
				T{       -1 0= -> <FALSE> }T
				T{ MAX-UINT 0= -> <FALSE> }T
				T{ MIN-INT  0= -> <FALSE> }T
				T{ MAX-INT  0= -> <FALSE> }T
			`,
		},

		{
			name: "0<",
			code: `
				T{       0 0< -> <FALSE> }T
				T{      -1 0< -> <TRUE>  }T
				T{ MIN-INT 0< -> <TRUE>  }T
				T{       1 0< -> <FALSE> }T
				T{ MAX-INT 0< -> <FALSE> }T
			`,
		},
		{
			name: "0>",
			code: `
				T{       0 0> -> <FALSE> }T
				T{      -1 0> -> <FALSE> }T
				T{ MIN-INT 0> -> <FALSE> }T
				T{       1 0> -> <TRUE>  }T
				T{ MAX-INT 0> -> <TRUE>  }T
			`,
		},
		{
			name: "0<>",
			code: `
				T{        0 0<> -> <FALSE> }T
				T{        1 0<> -> <TRUE>  }T
				T{        2 0<> -> <TRUE>  }T
				T{       -1 0<> -> <TRUE>  }T
				T{ MAX-UINT 0<> -> <TRUE>  }T
				T{ MIN-INT  0<> -> <TRUE>  }T
				T{ MAX-INT  0<> -> <TRUE>  }T
			`,
		},
		{
			name: "<",
			code: `
				T{       0       1 < -> <TRUE>  }T
				T{       1       2 < -> <TRUE>  }T
				T{      -1       0 < -> <TRUE>  }T
				T{      -1       1 < -> <TRUE>  }T
				T{ MIN-INT       0 < -> <TRUE>  }T
				T{ MIN-INT MAX-INT < -> <TRUE>  }T
				T{       0 MAX-INT < -> <TRUE>  }T
				T{       0       0 < -> <FALSE> }T
				T{       1       1 < -> <FALSE> }T
				T{       1       0 < -> <FALSE> }T
				T{       2       1 < -> <FALSE> }T
				T{       0      -1 < -> <FALSE> }T
				T{       1      -1 < -> <FALSE> }T
				T{       0 MIN-INT < -> <FALSE> }T
				T{ MAX-INT MIN-INT < -> <FALSE> }T
				T{ MAX-INT       0 < -> <FALSE> }T
			`,
		},
		{
			name: ">",
			code: `
				T{       0       1 > -> <FALSE> }T
				T{       1       2 > -> <FALSE> }T
				T{      -1       0 > -> <FALSE> }T
				T{      -1       1 > -> <FALSE> }T
				T{ MIN-INT       0 > -> <FALSE> }T
				T{ MIN-INT MAX-INT > -> <FALSE> }T
				T{       0 MAX-INT > -> <FALSE> }T
				T{       0       0 > -> <FALSE> }T
				T{       1       1 > -> <FALSE> }T
				T{       1       0 > -> <TRUE>  }T
				T{       2       1 > -> <TRUE>  }T
				T{       0      -1 > -> <TRUE>  }T
				T{       1      -1 > -> <TRUE>  }T
				T{       0 MIN-INT > -> <TRUE>  }T
				T{ MAX-INT MIN-INT > -> <TRUE>  }T
				T{ MAX-INT       0 > -> <TRUE>  }T
			`,
		},
	}

	r := asm.Runner{}
	r.SetDefaults()
	err := r.SetupPort()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := createTest(tt.code)
			runOutputTest(code, "", t, &r)
		})
	}
}

func createTest(code string) string {
	return fmt.Sprintf("%s : MAIN %s ESP.DONE ; ", suite, code)
}
