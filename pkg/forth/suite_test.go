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
		name  string // the name of the test
		setup string // any code that we want to execute outside of the "main" word
		code  string // any code we want to run in the "main" word
		// tests that run in the global context should be added to suite_test.f
	}{
		// Core tests
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
			setup: `
				T{ DEFER defer1 -> }T
				T{ : action-defer1 ACTION-OF defer1 ; -> }T
				T{ ' * ' defer1 DEFER! ->   }T
				T{          2 3 defer1 -> 6 }T
				T{ ACTION-OF defer1 -> ' * }T
				T{    action-defer1 -> ' * }T
				T{ ' + IS defer1 ->   }T
				T{    1 2 defer1 -> 3 }T
				T{ ACTION-OF defer1 -> ' + }T
				T{    action-defer1 -> ' + }T
			`,
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
			setup: `
				T{ : GN2 ( -- 16 10 )
    			BASE @ >R HEX BASE @ DECIMAL BASE @ R> BASE ! ; -> }T
			`,
			code: "T{ GN2 -> 10 A }T",
		},
		// BEGIN doesn't have tests
		{
			name: "BL",
			code: "T{ BL -> 20 }T",
		},
		{
			name: "BUFFER:",
			setup: `
				DECIMAL
				T{ 127 CHARS BUFFER: TBUF1 -> }T
				T{ 127 CHARS BUFFER: TBUF2 -> }T \ Buffer is aligned
				T{ TBUF1 ALIGNED -> TBUF1 }T \ Buffers do not overlap
				\ T{ TBUF2 TBUF1 - ABS 127 CHARS < -> <FALSE> }T \ Buffer can be written to
				\ 1 CHARS CONSTANT /CHAR
				\ : TFULL? ( c-addr n char -- flag )
				\    TRUE 2SWAP CHARS OVER + SWAP ?DO
				\      OVER I C@ = AND
				\    /CHAR +LOOP NIP
				\ ;
				\ T{ TBUF1 127 CHAR * FILL   ->        }T
				\ T{ TBUF1 127 CHAR * TFULL? -> <TRUE> }T
				\ T{ TBUF1 127 0 FILL   ->        }T
				\ T{ TBUF1 127 0 TFULL? -> <TRUE> }T
			`,
			code: `
				T{ TBUF1 ALIGNED -> TBUF1 }T \ Buffers do not overlap
				\ T{ TBUF2 TBUF1 - ABS 127 CHARS < -> <FALSE> }T \ Buffer can be written to
				\ T{ TBUF1 127 CHAR * FILL   ->        }T
				\ T{ TBUF1 127 CHAR * TFULL? -> <TRUE> }T
				\ T{ TBUF1 127 0 FILL   ->        }T
				\ T{ TBUF1 127 0 TFULL? -> <TRUE> }T
			`,
		},
		{
			name:  "[",
			setup: "T{ : GC3 [ GC1 ] LITERAL ; -> }T",
			code:  "T{ GC3 -> 58 }T",
		},
		{
			name: "[CHAR]",
			setup: `
				T{ : GC2 [CHAR] HELLO ; -> }T
			`,
			code: `
				T{ GC1 -> 58 }T
				T{ GC2 -> 48 }T
			`,
		},
		{
			name: "[COMPILE]",
			setup: `
				T{ : [c1] [COMPILE] DUP ; IMMEDIATE -> }T
				T{ : [c2] [COMPILE] [c1] ; -> }T
				T{ : [cif] [COMPILE] IF ; IMMEDIATE -> }T
				T{ : [c3]  [cif] 111 ELSE 222 THEN ; -> }T
			`,
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
			setup: `
				: cs1 CASE 1 OF 111 ENDOF
					2 OF 222 ENDOF
					3 OF 333 ENDOF
					>R 999 R>
					ENDCASE
				;
				: cs2 >R CASE
				-1 OF CASE R@ 1 OF 100 ENDOF
								2 OF 200 ENDOF
								>R -300 R>
						ENDCASE
					ENDOF
				-2 OF CASE R@ 1 OF -99 ENDOF
								>R -199 R>
						ENDCASE
					ENDOF
					>R 299 R>
				ENDCASE R> DROP ;
			`,
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
			setup: `
				T{ 123 CONSTANT X123 -> }T
				T{ : EQU CONSTANT ; -> }T
				T{ X123 EQU Y123 -> }T
			`,
			code: `
				T{ X123 -> 123 }T
				T{ Y123 -> 123 }T
			`,
		},
		{
			name:  "COUNT",
			setup: "T{ BL WORD GT1 CONSTANT GT1STRING -> }T", // modified from FIND test
			code: `
				T{ GT1STRING COUNT -> GT1STRING CHAR+ 3 }T
			`,
		},
		// CR does not have any tests
		// CREATE does not have any tests
		// C! does not have any tests
		{
			name: ":",
			setup: `
				T{ : NOP : POSTPONE ; ; -> }T
				T{ NOP NOP1 NOP NOP2 -> }T
				T{ : GDX 123 ; : GDX GDX 234 ; -> }T
			`,
			code: `
				T{ NOP1 -> }T
				T{ NOP2 -> }T
				T{ GDX -> 123 234 }T
			`,
		},
		{
			name: ":NONAME",
			setup: `
				T{ VARIABLE nn1 -> }T
				T{ VARIABLE nn2 -> }T
				T{ :NONAME 1234 ; nn1 ! -> }T
				T{ :NONAME 9876 ; nn2 ! -> }T
			`,
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
			setup: `
				T{ DEFER defer2 -> }T
				T{ ' * ' defer2 DEFER! -> }T
				T{   2 3 defer2 -> 6 }T
				T{ ' + IS defer2 ->   }T
				T{    1 2 defer2 -> 3 }T
			`,
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
			setup: `
				T{ DEFER defer4 -> }T
				T{ ' * ' defer4 DEFER! -> }T
				T{ 2 3 defer4 -> 6 }T
				T{ ' defer4 DEFER@ -> ' * }T
				T{ ' + IS defer4 -> }T
				T{ 1 2 defer4 -> 3 }T
				T{ ' defer4 DEFER@ -> ' + }T
			`,
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
			setup: `
				T{ DEFER defer3 -> }T
				T{ ' * ' defer3 DEFER! -> }T
				T{ 2 3 defer3 -> 6 }T
				T{ ' + ' defer3 DEFER! -> }T
				T{ 1 2 defer3 -> 3 }T
			`,
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
		{
			name: "/MOD",
			code: ` \ modified from test suite
				T{ 0 1 /MOD -> 0 0 }T
				T{ 1 1 /MOD -> 0 1 }T
				T{ 2 1 /MOD -> 0 2 }T
				T{ -1 1 /MOD -> 0 -1 }T
				T{ -2 1 /MOD -> 0 -2 }T
				T{ 0 -1 /MOD -> 0 0 }T
				T{ 1 -1 /MOD -> 0 -1 }T
				T{ 2 -1 /MOD -> 0 -2 }T
				T{ -1 -1 /MOD -> 0 1 }T
				T{ -2 -1 /MOD -> 0 2 }T
				T{ 2 2 /MOD -> 0 1 }T
				T{ -2 -2 /MOD -> 0 1 }T
				T{ 7 3 /MOD -> 1 2 }T
				T{ 7 -3 /MOD -> -2 -3 }T
			`,
		},
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
			setup: `
				T{ : GI1 IF 123 THEN ; -> }T
				T{ : GI2 IF 123 ELSE 234 THEN ; -> }T
				\ Multiple ELSEs in an IF statement
				T{ : melse IF 1 ELSE 2 ELSE 3 ELSE 4 ELSE 5 THEN ; -> }T
			`,
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
			setup: `
				T{ 123 CONSTANT iw1 IMMEDIATE iw1 -> 123 }T
				T{ : iw2 iw1 LITERAL ; iw2 -> 123 }T
				T{ VARIABLE iw3 IMMEDIATE 234 iw3 ! iw3 @ -> 234 }T
				T{ : iw4 iw3 [ @ ] LITERAL ; iw4 -> 234 }T
				T{ :NONAME [ 345 ] iw3 [ ! ] ; DROP iw3 @ -> 345 }T
				\ T{ CREATE iw5 456 , IMMEDIATE -> }T
				\ T{ :NONAME iw5 [ @ iw3 ! ] ; DROP iw3 @ -> 456 }T
				\ T{ : iw6 CREATE , IMMEDIATE DOES> @ 1+ ; -> }T
				\ T{ 111 iw6 iw7 iw7 -> 112 }T
				\ T{ : iw8 iw7 LITERAL 1+ ; iw8 -> 113 }T
				\ T{ : iw9 CREATE , DOES> @ 2 + IMMEDIATE ; -> }T
				\ : find-iw BL WORD FIND NIP ;
				\ T{ 222 iw9 iw10 find-iw iw10 -> -1 }T    \ iw10 is not immediate
				\ T{ iw10 find-iw iw10 -> 224 1 }T          \ iw10 becomes immediate
			`,
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
			setup: `
				T{ DEFER defer5 -> }T
				T{ : is-defer5 IS defer5 ; -> }T
				T{ ' * IS defer5 -> }T
				T{ 2 3 defer5 -> 6 }T
				T{ ' + is-defer5 -> }T
				T{ 1 2 defer5 -> 3 }T
			`,
			code: `
				\ uses ['] rather than ' so we can run on ulp
				T{ ['] * IS defer5 -> }T
				T{ 2 3 defer5 -> 6    }T

				T{ ['] + is-defer5 -> }T
				T{ 1 2 defer5 -> 3    }T
			`,
		},
		{
			name: "J",
			setup: `
				T{ : GD3 DO 1 0 DO J LOOP LOOP ; -> }T
				T{ : GD4 DO 1 0 DO J LOOP -1 +LOOP ; -> }T
			`,
			code: `
				T{          4        1 GD3 ->  1 2 3   }T
				T{          2       -1 GD3 -> -1 0 1   }T
				T{ MID-UINT+1 MID-UINT GD3 -> MID-UINT }T
				T{        1          4 GD4 -> 4 3 2 1             }T
				T{       -1          2 GD4 -> 2 1 0 -1            }T
				T{ MID-UINT MID-UINT+1 GD4 -> MID-UINT+1 MID-UINT }T
			`,
		},
		// KEY
		{
			name: "LEAVE",
			setup: `
				T{ : GD5 123 SWAP 0 DO 
					I 4 > IF DROP 234 LEAVE THEN 
					LOOP ; -> }T
			`,
			code: `
				T{ 1 GD5 -> 123 }T
				T{ 5 GD5 -> 123 }T
				T{ 6 GD5 -> 234 }T
			`,
		},
		{
			name: "LITERAL",
			setup: `
				T{ : GT3 GT2 LITERAL ; -> }T
				T{ GT3 -> ' GT1 }T
			`,
			code: `
				T{ GT3 -> ['] GT1 }T
			`,
		},
		{
			name:  "LOOP",
			setup: "T{ : GD1 DO I LOOP ; -> }T",
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
			setup: `
				T{ : GT4 POSTPONE GT1 ; IMMEDIATE -> }T
				T{ : GT5 GT4 ; -> }T
				T{ : GT6 345 ; IMMEDIATE -> }T
				T{ : GT7 POSTPONE GT6 ; -> }T
			`,
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
		{
			name: "+LOOP",
			setup: `
				DECIMAL \ these tests run in DECIMAL mode
				T{ : GD2 DO I -1 +LOOP ; -> }T
				VARIABLE gditerations
				VARIABLE gdincrement
				: gd7 ( limit start increment -- )
					gdincrement !
					0 gditerations !
					DO
						1 gditerations +!
						I
						gditerations @ 6 = IF LEAVE THEN
						gdincrement @
					+LOOP gditerations @
				;
				MAX-UINT 8 RSHIFT 1+ CONSTANT ustep
				ustep NEGATE CONSTANT -ustep
				MAX-INT 7 RSHIFT 1+ CONSTANT step
				step NEGATE CONSTANT -step
				VARIABLE bump
				T{  : gd8 bump ! DO 1+ bump @ +LOOP ; -> }T
			`,
			code: `
				T{        1          4 GD2 -> 4 3 2  1 }T
				T{       -1          2 GD2 -> 2 1 0 -1 }T
				T{    4  4  -1 gd7 ->  4                  1  }T
				T{    1  4  -1 gd7 ->  4  3  2  1         4  }T
				T{    4  1  -1 gd7 ->  1  0 -1 -2  -3  -4 6  }T
				T{    4  1   0 gd7 ->  1  1  1  1   1   1 6  }T
				T{    0  0   0 gd7 ->  0  0  0  0   0   0 6  }T
				T{    1  4   0 gd7 ->  4  4  4  4   4   4 6  }T
				T{    1  4   1 gd7 ->  4  5  6  7   8   9 6  }T
				T{    4  1   1 gd7 ->  1  2  3            3  }T
				T{    4  4   1 gd7 ->  4  5  6  7   8   9 6  }T
				T{    2 -1  -1 gd7 -> -1 -2 -3 -4  -5  -6 6  }T
				T{   -1  2  -1 gd7 ->  2  1  0 -1         4  }T
				T{    2 -1   0 gd7 -> -1 -1 -1 -1  -1  -1 6  }T
				T{   -1  2   0 gd7 ->  2  2  2  2   2   2 6  }T
				T{   -1  2   1 gd7 ->  2  3  4  5   6   7 6  }T
				T{    2 -1   1 gd7 -> -1 0 1              3  }T
				T{  -20 30 -10 gd7 -> 30 20 10  0 -10 -20 6  }T
				T{  -20 31 -10 gd7 -> 31 21 11  1  -9 -19 6  }T
				T{  -20 29 -10 gd7 -> 29 19  9 -1 -11     5  }T
				T{  0 MAX-UINT 0 ustep gd8 -> 256 }T
				T{  0 0 MAX-UINT -ustep gd8 -> 256 }T
				T{  0 MAX-INT MIN-INT step gd8 -> 256 }T
				T{  0 MIN-INT MAX-INT -step gd8 -> 256 }T
			`,
		},
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
			setup: `
				DECIMAL
				T{ : GI6 ( N -- 0,1,..N ) 
					DUP IF DUP >R 1- RECURSE R> THEN ; -> }T
				T{ :NONAME ( n -- 0, 1, .., n ) 
					DUP IF DUP >R 1- RECURSE R> THEN 
				;
				CONSTANT rn1 -> }T
				:NONAME ( n -- n1 )
				1- DUP
				CASE 0 OF EXIT ENDOF
					1 OF 11 SWAP RECURSE ENDOF
					2 OF 22 SWAP RECURSE ENDOF
					3 OF 33 SWAP RECURSE ENDOF
					DROP ABS RECURSE EXIT
				ENDCASE
				;
				CONSTANT rn2
			`,
			code: `
				T{ 0 GI6 -> 0 }T
				T{ 1 GI6 -> 0 1 }T
				T{ 2 GI6 -> 0 1 2 }T
				T{ 3 GI6 -> 0 1 2 3 }T
				T{ 4 GI6 -> 0 1 2 3 4 }T
				T{ 0 rn1 EXECUTE -> 0 }T
				T{ 4 rn1 EXECUTE -> 0 1 2 3 4 }T
				T{  1 rn2 EXECUTE -> 0 }T
				T{  2 rn2 EXECUTE -> 11 0 }T
				T{  4 rn2 EXECUTE -> 33 22 11 0 }T
				T{ 25 rn2 EXECUTE -> 33 22 11 0 }T
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
			setup: `
				T{ : GT8 STATE @ ; IMMEDIATE -> }T
				T{ : GT9 GT8 LITERAL ; -> }T
			`,
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
			setup: `
				T{ : GC4 S" XY" ; ->   }T
				T{ : GC5 S" A String"2DROP ; -> }T \ There is no space between the " and 2DROP
			`,
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
			setup: `
				T{ ' GT1 EXECUTE -> 123 }T
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
		{
			name: "UNLOOP",
			setup: `
				T{ : GD6 ( PAT: {0 0},{0 0}{1 0}{1 1},{0 0}{1 0}{1 1}{2 0}{2 1}{2 2} ) 
					0 SWAP 0 DO 
						I 1+ 0 DO 
						I J + 3 = IF I UNLOOP I UNLOOP EXIT THEN 1+ 
						LOOP 
					LOOP ; -> }T
			`,
			code: `
				T{ 1 GD6 -> 1 }T
				T{ 2 GD6 -> 3 }T
				T{ 3 GD6 -> 4 1 2 }T
			`,
		},
		{
			name:  "UNTIL",
			setup: "T{ : GI4 BEGIN DUP 1+ DUP 5 > UNTIL ; -> }T",
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
			name:  "VARIABLE",
			setup: "T{ VARIABLE V1 ->     }T",
			code: `
				T{    123 V1 ! ->     }T
				T{        V1 @ -> 123 }T
			`,
		},
		// VARIABLE
		{
			name: "WHILE",
			setup: `
				T{ : GI3 BEGIN DUP 5 < WHILE DUP 1+ REPEAT ; -> }T
				T{ : GI5 BEGIN DUP 2 > WHILE
					DUP 5 < WHILE DUP 1+ REPEAT
					123 ELSE 345 THEN ; -> }T
			`,
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
			name:  "\\",
			setup: "T{ : COMMENT POSTPONE \\ ; IMMEDIATE -> }T",
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
			name:  "(",
			setup: "T{ : pc1 ( A comment)1234 ; -> }T",
			code: `
				\ There is no space either side of the ).
				T{ ( A comment)1234 -> 1234 }T
				T{ pc1 -> 1234 }T
			`,
		},
		{
			name: "?DO",
			setup: `
				DECIMAL
				: qd ?DO I LOOP ;
				: qd1 ?DO I 10 +LOOP ;
				: qd2 ?DO I 3 > IF LEAVE ELSE I THEN LOOP ;
				: qd3 ?DO I 1 +LOOP ;
				: qd4 ?DO I -1 +LOOP ;
				: qd5 ?DO I -10 +LOOP ;
				VARIABLE qditerations
				VARIABLE qdincrement
				: qd6 ( limit start increment -- )    qdincrement !
					0 qditerations !
					?DO
						1 qditerations +!
						I
						qditerations @ 6 = IF LEAVE THEN
						qdincrement @
					+LOOP qditerations @
				;
			`,
			code: `
				T{   789   789 qd -> }T
				T{ -9876 -9876 qd -> }T
				T{     5     0 qd -> 0 1 2 3 4 }T
				T{ 50 1 qd1 -> 1 11 21 31 41 }T
				T{ 50 0 qd1 -> 0 10 20 30 40 }T
				T{ 5 -1 qd2 -> -1 0 1 2 3 }T
				T{ 4  4 qd3 -> }T
				T{ 4  1 qd3 ->  1 2 3 }T
				T{ 2 -1 qd3 -> -1 0 1 }T
				T{  4 4 qd4 -> }T
				T{  1 4 qd4 -> 4 3 2  1 }T
				T{ -1 2 qd4 -> 2 1 0 -1 }T
				T{   1 50 qd5 -> 50 40 30 20 10   }T
				T{   0 50 qd5 -> 50 40 30 20 10 0 }T
				T{ -25 10 qd5 -> 10 0 -10 -20     }T
				T{  4  4 -1 qd6 ->                   0  }T
				T{  1  4 -1 qd6 ->  4  3  2  1       4  }T
				T{  4  1 -1 qd6 ->  1  0 -1 -2 -3 -4 6  }T
				T{  4  1  0 qd6 ->  1  1  1  1  1  1 6  }T
				T{  0  0  0 qd6 ->                   0  }T
				T{  1  4  0 qd6 ->  4  4  4  4  4  4 6  }T
				T{  1  4  1 qd6 ->  4  5  6  7  8  9 6  }T
				T{  4  1  1 qd6 ->  1  2  3          3  }T
				T{  4  4  1 qd6 ->                   0  }T
				T{  2 -1 -1 qd6 -> -1 -2 -3 -4 -5 -6 6  }T
				T{ -1  2 -1 qd6 ->  2  1  0 -1       4  }T
				T{  2 -1  0 qd6 -> -1 -1 -1 -1 -1 -1 6  }T
				T{ -1  2  0 qd6 ->  2  2  2  2  2  2 6  }T
				T{ -1  2  1 qd6 ->  2  3  4  5  6  7 6  }T
				T{  2 -1  1 qd6 -> -1  0  1          3  }T
			`,
		},
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
			setup: `
				T{ : GR1 >R R> ; -> }T
				T{ : GR2 >R R@ R> DROP ; -> }T
			`,
			code: `
				T{ 123 GR1 -> 123 }T
				T{ 123 GR2 -> 123 }T
				T{  1S GR1 ->  1S }T      ( Return stack holds cells )
			`,
		},

		// Double tests
		// DABS
		// D.R
		{
			name: "D=",
			code: `
				T{      -1.      -1. D= -> <TRUE>  }T
				T{      -1.       0. D= -> <FALSE> }T
				T{      -1.       1. D= -> <FALSE> }T
				T{       0.      -1. D= -> <FALSE> }T
				T{       0.       0. D= -> <TRUE>  }T
				T{       0.       1. D= -> <FALSE> }T
				T{       1.      -1. D= -> <FALSE> }T
				T{       1.       0. D= -> <FALSE> }T
				T{       1.       1. D= -> <TRUE>  }T
				T{   0   -1    0  -1 D= -> <TRUE>  }T
				T{   0   -1    0   0 D= -> <FALSE> }T
				T{   0   -1    0   1 D= -> <FALSE> }T
				T{   0    0    0  -1 D= -> <FALSE> }T
				T{   0    0    0   0 D= -> <TRUE>  }T
				T{   0    0    0   1 D= -> <FALSE> }T
				T{   0    1    0  -1 D= -> <FALSE> }T
				T{   0    1    0   0 D= -> <FALSE> }T
				T{   0    1    0   1 D= -> <TRUE>  }T
				T{ MAX-2INT MIN-2INT D= -> <FALSE> }T
				T{ MAX-2INT       0. D= -> <FALSE> }T
				T{ MAX-2INT MAX-2INT D= -> <TRUE>  }T
				T{ MAX-2INT HI-2INT  D= -> <FALSE> }T
				T{ MAX-2INT MIN-2INT D= -> <FALSE> }T
				T{ MIN-2INT MIN-2INT D= -> <TRUE>  }T
				T{ MIN-2INT LO-2INT  D= -> <FALSE> }T
				T{ MIN-2INT MAX-2INT D= -> <FALSE> }T
			`,
		},
		// DMAX
		// DMIN
		// D-
		// DNEGATE
		// D+
		// D2/
		// D2*
		// DU<
		// D0=
		// D0<
		// D.
		// D>S
		// M+
		// M*/
		// 2CONSTANT
		// 2LITERAL
		// 2ROT
		// 2VALUE
		// 2VARIABLE
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
			code := createTest(tt.setup, tt.code)
			runOutputTest(code, "DONE", t, &r)
		})
	}
}

func createTest(setup, code string) string {
	return fmt.Sprintf("%s %s : MAIN %s .\" DONE\" ESP.DONE ; ", suite, setup, code)
}
