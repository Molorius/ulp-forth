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
		// AGAIN does not have test cases
		// ALIGN
		// ALIGNED does not have test cases
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
		{
			name: "BUFFER:",
			code: `
				DECIMAL
				T{ TBUF1 ALIGNED -> TBUF1 }T \ Buffers do not overlap
				\ T{ TBUF2 TBUF1 - ABS 127 CHARS < -> <FALSE> }T \ Buffer can be written to
				\ T{ TBUF1 127 CHAR * FILL   ->        }T
				\ T{ TBUF1 127 CHAR * TFULL? -> <TRUE> }T
				\ T{ TBUF1 127 0 FILL   ->        }T
				\ T{ TBUF1 127 0 TFULL? -> <TRUE> }T
			`,
		},
		{
			name: "[",
			code: "T{ GC3 -> 58 }T",
		},
		{
			name: "[CHAR]",
			code: `
				T{ GC1 -> 58 }T
				T{ GC2 -> 48 }T
			`,
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
		{
			name: "CASE",
			code: `
				T{ 1 cs1 -> 111 }T
				T{ 2 cs1 -> 222 }T
				T{ 3 cs1 -> 333 }T
				T{ 4 cs1 -> 999 }T
				T{ -1 1 cs2 ->  100 }T
				T{ -1 2 cs2 ->  200 }T
				T{ -1 3 cs2 -> -300 }T
				T{ -2 1 cs2 ->  -99 }T
				T{ -2 2 cs2 -> -199 }T
				T{  0 2 cs2 ->  299 }T
			`,
		},
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
		// /
		// /MOD
		// .R
		// .(
		// ."
		// ELSE does not have any tests
		// EMIT testing should be done in basic_test.go
		// ENDCASE does not have any tests
		// ENDOF does not have any tests
		// ENVIRONMENT?
		// ERASE
		// EVALUATE
		// EXECUTE does not have any tests
		// EXIT does not have any tests
		{
			name: "=",
			code: `
				T{  0  0 = -> <TRUE>  }T
				T{  1  1 = -> <TRUE>  }T
				T{ -1 -1 = -> <TRUE>  }T
				T{  1  0 = -> <FALSE> }T
				T{ -1  0 = -> <FALSE> }T
				T{  0  1 = -> <FALSE> }T
				T{  0 -1 = -> <FALSE> }T
			`,
		},
		{
			name: "FALSE",
			code: `
				T{ FALSE -> 0 }T
				T{ FALSE -> <FALSE> }T
			`,
		},
		// FILL
		// FIND
		// FM/MOD
		// @ does not have any tests
		// HERE
		// HEX does not have any tests
		// HOLD
		// HOLDS
		// I
		// IF
		{
			name: "IF",
			code: `
				T{  0 GI1 ->     }T
				T{  1 GI1 -> 123 }T
				T{ -1 GI1 -> 123 }T
				T{  0 GI2 -> 234 }T
				T{  1 GI2 -> 123 }T
				T{ -1 GI1 -> 123 }T
				T{ <FALSE> melse -> 2 4 }T
				T{ <TRUE>  melse -> 1 3 5 }T
			`,
		},
		{
			name: "IMMEDIATE",
			code: `
				\ T{ 222 iw9 iw10 find-iw iw10 -> -1 }T    \ iw10 is not immediate
				\ T{ iw10 find-iw iw10 -> 224 1 }T          \ iw10 becomes immediate
			`,
		},
		{
			name: "INVERT",
			code: `
				T{ 0S INVERT -> 1S }T
				T{ 1S INVERT -> 0S }T
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
		// J
		// KEY
		// LEAVE
		{
			name: "LITERAL",
			code: `
				T{ GT3 -> ['] GT1 }T
			`,
		},
		{
			name: "LOOP",
			code: `
				T{          4        1 GD1 ->  1 2 3   }T
				T{          2       -1 GD1 -> -1 0 1   }T
				T{ MID-UINT+1 MID-UINT GD1 -> MID-UINT }T
			`,
		},
		{
			name: "LSHIFT",
			code: `
				T{   1 0 LSHIFT ->    1 }T
				T{   1 1 LSHIFT ->    2 }T
				T{   1 2 LSHIFT ->    4 }T
				T{   1 F LSHIFT -> 8000 }T      \ BIGGEST GUARANTEED SHIFT
				T{  1S 1 LSHIFT 1 XOR -> 1S }T
				T{ MSB 1 LSHIFT ->    0 }T
			`,
		},
		// MARKER
		{
			name: "MAX",
			code: `
				T{       0       1 MAX ->       1 }T
				T{       1       2 MAX ->       2 }T
				T{      -1       0 MAX ->       0 }T
				T{      -1       1 MAX ->       1 }T
				T{ MIN-INT       0 MAX ->       0 }T
				T{ MIN-INT MAX-INT MAX -> MAX-INT }T
				T{       0 MAX-INT MAX -> MAX-INT }T
				T{       0       0 MAX ->       0 }T
				T{       1       1 MAX ->       1 }T
				T{       1       0 MAX ->       1 }T
				T{       2       1 MAX ->       2 }T
				T{       0      -1 MAX ->       0 }T
				T{       1      -1 MAX ->       1 }T
				T{       0 MIN-INT MAX ->       0 }T
				T{ MAX-INT MIN-INT MAX -> MAX-INT }T
				T{ MAX-INT       0 MAX -> MAX-INT }T
			`,
		},
		{
			name: "MIN",
			code: `
				T{       0       1 MIN ->       0 }T
				T{       1       2 MIN ->       1 }T
				T{      -1       0 MIN ->      -1 }T
				T{      -1       1 MIN ->      -1 }T
				T{ MIN-INT       0 MIN -> MIN-INT }T
				T{ MIN-INT MAX-INT MIN -> MIN-INT }T
				T{       0 MAX-INT MIN ->       0 }T
				T{       0       0 MIN ->       0 }T
				T{       1       1 MIN ->       1 }T
				T{       1       0 MIN ->       0 }T
				T{       2       1 MIN ->       1 }T
				T{       0      -1 MIN ->      -1 }T
				T{       1      -1 MIN ->      -1 }T
				T{       0 MIN-INT MIN -> MIN-INT }T
				T{ MAX-INT MIN-INT MIN -> MIN-INT }T
				T{ MAX-INT       0 MIN ->       0 }T
			`,
		},
		// MOD
		// MOVE
		// M*
		{
			name: "-",
			code: `
				T{          0  5 - ->       -5 }T
				T{          5  0 - ->        5 }T
				T{          0 -5 - ->        5 }T
				T{         -5  0 - ->       -5 }T
				T{          1  2 - ->       -1 }T
				T{          1 -2 - ->        3 }T
				T{         -1  2 - ->       -3 }T
				T{         -1 -2 - ->        1 }T
				T{          0  1 - ->       -1 }T
				T{ MID-UINT+1  1 - -> MID-UINT }T
			`,
		},
		{
			name: "NEGATE",
			code: `
				T{  0 NEGATE ->  0 }T
				T{  1 NEGATE -> -1 }T
				T{ -1 NEGATE ->  1 }T
				T{  2 NEGATE -> -2 }T
				T{ -2 NEGATE ->  2 }T
			`,
		},
		{
			name: "NIP",
			code: `
				T{ 1 2 NIP -> 2 }T
			`,
		},
		// OF does not have any tests
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
			name: "OVER",
			code: `
				T{ 1 2 OVER -> 1 2 1 }T
			`,
		},
		{
			name: "1-",
			code: `
				T{          2 1- ->        1 }T
				T{          1 1- ->        0 }T
				T{          0 1- ->       -1 }T
				T{ MID-UINT+1 1- -> MID-UINT }T
			`,
		},
		{
			name: "1+",
			code: `
				T{        0 1+ ->          1 }T
				T{       -1 1+ ->          0 }T
				T{        1 1+ ->          2 }T
				T{ MID-UINT 1+ -> MID-UINT+1 }T
			`,
		},
		// PAD
		// PARSE-NAME
		// PARSE
		{
			name: "PICK",
			code: `
				T{ 123 456 789 2 PICK -> 123 456 789 123 }T
			`,
		},
		{
			name: "POSTPONE",
			code: `
				T{ GT5 -> 123 }T
				T{ GT7 -> 345 }T
			`,
		},
		{
			name: "+",
			code: `
				T{        0  5 + ->          5 }T
				T{        5  0 + ->          5 }T
				T{        0 -5 + ->         -5 }T
				T{       -5  0 + ->         -5 }T
				T{        1  2 + ->          3 }T
				T{        1 -2 + ->         -1 }T
				T{       -1  2 + ->          1 }T
				T{       -1 -2 + ->         -3 }T
				T{       -1  1 + ->          0 }T
				T{ MID-UINT  1 + -> MID-UINT+1 }T
			`,
		},
		// +LOOP
		{
			name: "+!",
			code: `
				T{  0 1ST !        ->   }T
				T{  1 1ST +!       ->   }T
				T{    1ST @        -> 1 }T
				T{ -1 1ST +! 1ST @ -> 0 }T
			`,
		},
		// QUIT
		{
			name: "RECURSE",
			code: `
				T{ 0 GI6 -> 0 }T
				T{ 1 GI6 -> 0 1 }T
				T{ 2 GI6 -> 0 1 2 }T
				T{ 3 GI6 -> 0 1 2 3 }T
				T{ 4 GI6 -> 0 1 2 3 4 }T
				DECIMAL
				T{ 0 rn1 EXECUTE -> 0 }T
				T{ 4 rn1 EXECUTE -> 0 1 2 3 4 }T
			`,
		},
		// REFILL
		// REPEAT does not have test cases
		// RESTORE-INPUT
		// R@ does not have test cases
		// ROLL
		{
			name: "ROT",
			code: `
				T{ 1 2 3 ROT -> 2 3 1 }T
			`,
		},
		{
			name: "RSHIFT",
			code: `
				T{    1 0 RSHIFT -> 1 }T
				T{    1 1 RSHIFT -> 0 }T
				T{    2 1 RSHIFT -> 1 }T
				T{    4 2 RSHIFT -> 1 }T
				T{ 8000 F RSHIFT -> 1 }T                \ Biggest
				T{  MSB 1 RSHIFT MSB AND ->   0 }T    \ RSHIFT zero fills MSBs
				T{  MSB 1 RSHIFT     2*  -> MSB }T
			`,
		},
		// R> does not have test cases
		// SAVE-INPUT
		// SIGN
		// SM/REM
		// SOURCE-ID
		// SOURCE
		// SPACE
		// SPACES
		{
			name: "STATE",
			code: `
				T{ GT9 0= -> <FALSE> }T
			`,
		},
		{
			name: "SWAP",
			code: `
				T{ 1 2 SWAP -> 2 1 }T
			`,
		},
		// ; does not have test cases
		{
			name: "S\"",
			code: `
				T{ GC4 SWAP DROP  -> 2 }T
				T{ GC4 DROP DUP C@ SWAP CHAR+ C@ -> 58 59 }T
				T{ GC5 -> }T
			`,
		},
		// S"
		{
			name: "S>D",
			code: `
				T{       0 S>D ->       0  0 }T
				T{       1 S>D ->       1  0 }T
				T{       2 S>D ->       2  0 }T
				T{      -1 S>D ->      -1 -1 }T
				T{      -2 S>D ->      -2 -1 }T
				T{ MIN-INT S>D -> MIN-INT -1 }T
				T{ MAX-INT S>D -> MAX-INT  0 }T
			`,
		},
		// ! does not have test cases
		// THEN does not have test cases
		// TO
		{
			name: "TRUE",
			code: `
				T{ TRUE -> <TRUE> }T
				T{ TRUE -> 0 INVERT }T
			`,
		},
		{
			name: "TUCK",
			code: `
				T{ 1 2 TUCK -> 2 1 2 }T
			`,
		},
		// TYPE does not have test cases
		{
			name: "'",
			code: `
				\ T{ ['] GT1 EXECUTE -> 123 }T
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
		// */
		// */MOD
		{
			name: "2DROP",
			code: `
				T{ 1 2 2DROP -> }T
			`,
		},
		{
			name: "2DUP",
			code: `
				T{ 1 2 2DUP -> 1 2 1 2 }T
			`,
		},
		{
			name: "2/",
			code: `
				T{          0S 2/ ->   0S }T
				T{           1 2/ ->    0 }T
				T{        4000 2/ -> 2000 }T
				T{          1S 2/ ->   1S }T \ MSB PROPOGATED
				T{    1S 1 XOR 2/ ->   1S }T
				T{ MSB 2/ MSB AND ->  MSB }T
			`,
		},
		// 2@
		{
			name: "2OVER",
			code: `
				T{ 1 2 3 4 2OVER -> 1 2 3 4 1 2 }T
			`,
		},
		// 2R@
		// 2R> does not have any test cases
		// 2SWAP
		// 2!
		{
			name: "2*",
			code: `
				T{   0S 2*       ->   0S }T
				T{    1 2*       ->    2 }T
				T{ 4000 2*       -> 8000 }T
				T{   1S 2* 1 XOR ->   1S }T
				T{  MSB 2*       ->   0S }T
			`,
		},
		// 2>R does not have any test cases
		// U.R
		// UM/MOD
		// UM*
		// UNLOOP
		{
			name: "UNTIL",
			code: `
				T{ 3 GI4 -> 3 4 5 6 }T
				T{ 5 GI4 -> 5 6 }T
				T{ 6 GI4 -> 6 7 }T
			`,
		},
		// UNUSED
		// U. does not have any test cases
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
		// VALUE
		{
			name: "VARIABLE",
			code: `
				T{    123 V1 ! ->     }T
				T{        V1 @ -> 123 }T
			`,
		},
		// VARIABLE
		{
			name: "WHILE",
			code: `
				T{ 0 GI3 -> 0 1 2 3 4 5 }T
				T{ 4 GI3 -> 4 5 }T
				T{ 5 GI3 -> 5 }T
				T{ 6 GI3 -> 6 }T
				T{ 1 GI5 -> 1 345 }T
				T{ 2 GI5 -> 2 345 }T
				T{ 3 GI5 -> 3 4 5 123 }T
				T{ 4 GI5 -> 4 5 123 }T
				T{ 5 GI5 -> 5 123 }T
			`,
		},
		// WITHIN
		// WORD
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
			name: "\\",
			code: `
				T{ COMMENT \-postpone-fail
				-> }T
				\ \-inline-fail
			`,
		},
		// .
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
		{
			name: "<>",
			code: `
				T{  0  0 <> -> <FALSE>  }T
				T{  1  1 <> -> <FALSE>  }T
				T{ -1 -1 <> -> <FALSE>  }T
				T{  1  0 <> -> <TRUE> }T
				T{ -1  0 <> -> <TRUE> }T
				T{  0  1 <> -> <TRUE> }T
				T{  0 -1 <> -> <TRUE> }T
			`,
		},
		// #>
		// <#
		// #
		// #S
		{
			name: "(",
			code: `
				\ There is no space either side of the ).
				T{ ( A comment)1234 -> 1234 }T
			`,
		},
		// ?DO
		{
			name: "?DUP",
			code: `
				T{ -1 ?DUP -> -1 -1 }T
				T{  0 ?DUP ->  0    }T
				T{  1 ?DUP ->  1  1 }T
			`,
		},
		// >BODY does not have a test case we can replicate
		// >IN
		// >NUMBER
		{
			name: ">R",
			code: `
				T{ 123 GR1 -> 123 }T
				T{ 123 GR2 -> 123 }T
				T{  1S GR1 ->  1S }T      ( Return stack holds cells )
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
