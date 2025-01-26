/*
Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import (
	_ "embed"
	"fmt"
	"strings"
	"testing"

	"github.com/Molorius/ulp-c/pkg/asm"
)

//go:embed test/suite_test.f
var suite string

type suiteTest struct {
	name   string // the name of the test
	setup  string // any code that we want to execute outside of the "main" word
	code   string // any code we want to run in the "main" word
	expect string // any string we're expecting
	// tests that run in the global context should be added to suite_test.f
}

func TestCoreSuite(t *testing.T) {
	tests := []suiteTest{
		// ! see ,
		// # not implemented
		// #> not implemented
		// #S not implemented
		{
			name: "'",
			setup: `
				T{ ' GT1 EXECUTE -> 123 }T
			`,
		},
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
		// */ not implemented
		// */MOD not implemented
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
			name: "+!",
			code: `
				T{  0 1ST !        ->   }T
				T{  1 1ST +!       ->   }T
				T{    1ST @        -> 1 }T
				T{ -1 1ST +! 1ST @ -> 0 }T
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
			name: ",",
			code: `
				T{       1ST 2ND - 0< -> <TRUE> }T \ HERE MUST GROW WITH ALLOT
				T{       1ST CELL+    -> 2ND }T \ ... BY ONE CELL
				T{   1ST 1 CELLS +    -> 2ND }T
				T{     1ST @ 2ND @    -> 1 2 }T
				T{         5 1ST !    ->     }T
				T{     1ST @ 2ND @    -> 5 2 }T
				T{         6 2ND !    ->     }T
				T{     1ST @ 2ND @    -> 5 6 }T
				T{           1ST 2@   -> 6 5 }T
				T{       2 1 1ST 2!   ->     }T
				T{           1ST 2@   -> 2 1 }T
				T{ 1S 1ST !  1ST @    -> 1S  }T    \ CAN STORE CELL-WIDE VALUE
			`,
		},
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
		// . see EMIT
		{
			name: ".\"", // ."
			setup: `
				T{ : pb1 ." You should see 2345: "." 2345"; -> }T
			`,
			code: `
				T{ pb1 -> }T
			`,
			expect: "You should see 2345: 2345",
		},
		// / todo we need double divide to implement
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
			name: "1+",
			code: `
				T{        0 1+ ->          1 }T
				T{       -1 1+ ->          0 }T
				T{        1 1+ ->          2 }T
				T{ MID-UINT 1+ -> MID-UINT+1 }T
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
		// 2! see ,
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
		// 2@ see ,
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
			name: "2OVER",
			code: `
				T{ 1 2 3 4 2OVER -> 1 2 3 4 1 2 }T
			`,
		},
		{
			name: "2SWAP",
			code: "T{ 1 2 3 4 2SWAP -> 3 4 1 2 }T",
		},
		{
			name: ":",
			setup: `
				T{ : NOP : POSTPONE ; ; -> }T
				T{ NOP NOP1 NOP NOP2 -> }T
				T{ : GDX 123 ; : GDX2 GDX 234 ; -> }T
			`,
			code: `
				T{ NOP1 -> }T
				T{ NOP2 -> }T
				T{ GDX2 -> 123 234 }T
			`,
		},
		// ; see :
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
		// <# not implemented
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
			name: ">BODY",
			setup: `
				T{  CREATE CR0 ->      }T
				T{ ' CR0 >BODY -> HERE }T
			`,
		},
		// >IN not implemented
		// >NUMBER not implemented
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
		{
			name: "?DUP",
			code: `
				T{ -1 ?DUP -> -1 -1 }T
				T{  0 ?DUP ->  0    }T
				T{  1 ?DUP ->  1  1 }T
			`,
		},
		// @ see ,
		// ABORT not implemented
		// ABORT" not implemented
		{
			name: "ABS",
			code: `
				T{       0 ABS ->          0 }T
				T{       1 ABS ->          1 }T
				T{      -1 ABS ->          1 }T
				T{ MIN-INT ABS -> MID-UINT+1 }T
			`,
		},
		// ACCEPT not implemented
		{
			name: "ALIGN",
			setup: `
				ALIGN 1 ALLOT HERE ALIGN HERE 3 CELLS ALLOT
				CONSTANT A-ADDR CONSTANT UA-ADDR
				T{ UA-ADDR ALIGNED -> A-ADDR }T
			`,
			code: `
				T{       1 A-ADDR C!         A-ADDR       C@ ->       1 }T
				T{    1234 A-ADDR !          A-ADDR       @  ->    1234 }T
				T{ 123 456 A-ADDR 2!         A-ADDR       2@ -> 123 456 }T
				T{       2 A-ADDR CHAR+ C!   A-ADDR CHAR+ C@ ->       2 }T
				T{       3 A-ADDR CELL+ C!   A-ADDR CELL+ C@ ->       3 }T
				T{    1234 A-ADDR CELL+ !    A-ADDR CELL+ @  ->    1234 }T
				T{ 123 456 A-ADDR CELL+ 2!   A-ADDR CELL+ 2@ -> 123 456 }T
			`,
		},
		// ALIGNED does not have test cases
		{
			name: "ALLOT",
			setup: `
				HERE 1 ALLOT
				HERE
				CONSTANT 2NDA
				CONSTANT 1STA
			`,
			code: `
				T{ 1STA 2NDA - 0< -> <TRUE> }T    \ HERE MUST GROW WITH ALLOT
				T{      1STA 1+   ->   2NDA }T    \ ... BY ONE ADDRESS UNIT
			`,
		},
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
			name:  "C!", // not in test suite
			setup: "T{ VARIABLE c1 -> }T",
			code: `
				T{ 0x34 c1 C!       ->        }T
				T{ 0x12 c1 CHAR+ C! ->        }T
				T{ c1 @             -> 0x1234 }T
			`,
		},
		{
			name: "C,",
			setup: `
				HERE 1 C,
				HERE 2 C,
				CONSTANT 2NDC
				CONSTANT 1STC
			`,
			code: `
				T{    1STC 2NDC - 0< -> <TRUE> }T \ HERE MUST GROW WITH ALLOT
				T{      1STC CHAR+ ->  2NDC  }T \ ... BY ONE CHAR
				// T{  1STC 1 CHARS + ->  2NDC  }T \ this test is incorrect
				T{ 1STC C@ 2NDC C@ ->   1 2  }T
				T{       3 1STC C! ->        }T
				T{ 1STC C@ 2NDC C@ ->   3 2  }T
				T{       4 2NDC C! ->        }T
				T{ 1STC C@ 2NDC C@ ->   3 4  }T
			`,
		},
		{
			name:  "C@", // not in test suite
			setup: "T{ VARIABLE c1 -> }T",
			code: `
				T{ 0x1234 c1 ! ->      }T
				T{ c1 C@       -> 0x34 }T
				T{ c1 CHAR+ C@ -> 0x12 }T
			`,
		},
		// CELL+ see ,
		{
			name: "CELLS",
			setup: `
				: BITS ( X -- U )
					0 SWAP BEGIN DUP WHILE
					    DUP MSB AND IF >R 1+ R> THEN 2*
					REPEAT DROP ;
					( CELLS >= 1 AU, INTEGRAL MULTIPLE OF CHAR SIZE, >= 16 BITS )
			`,
			code: `
				T{ 1 CELLS 1 <         -> <FALSE> }T
				T{ 1 CELLS 1 CHARS MOD ->    0    }T
				T{ 1S BITS 10 <        -> <FALSE> }T
			`,
		},
		{
			name: "CHAR",
			setup: `
				T{ CHAR X     -> 58 }T
				T{ CHAR HELLO -> 48 }T
			`,
		},
		// CHAR+ see ,
		{
			name: "CHARS",
			code: `
				T{ 1 CHARS 1 <       -> <FALSE> }T
				T{ 1 CHARS 1 CELLS > -> <FALSE> }T
			`,
		},
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
		{
			name: "CR",
			code: `
				T{ CR -> }T
			`,
			expect: `
`,
		},
		{
			name: "CREATE",
			setup: `
				T{ CREATE CR1 -> }T
				T{ : ADDR1 [ HERE ] LITERAL ; -> }T
				T{ 1 , -> }T
				T{ CREATE CR2 -> }T
				T{ : ADDR2 [ HERE ] LITERAL ; -> }T
				T{ FF , -> }T
			`,
			code: `
				T{ CR1 -> ADDR1 }T
				T{ CR2 -> ADDR2 }T
				T{ CR1 @ -> 1  }T
				T{ CR2 @ -> FF }T
			`,
		},
		// DECIMAL see BASE
		{
			name: "DEPTH",
			code: `
				T{ 0 1 DEPTH -> 0 1 2 }T
				T{   0 DEPTH -> 0 1   }T
				T{     DEPTH -> 0     }T
			`,
		},
		// DO see LOOP +LOOP J I LEAVE UNLOOP
		{
			name: "DOES>",
			setup: `
				T{ : DOES1 DOES> @ 1 + ; -> }T
				T{ : DOES2 DOES> @ 2 + ; -> }T
				T{ CREATE CR1 -> }T
				T{ CR1   -> HERE }T
				T{ 1 ,   ->   }T
				T{ CR1 @ -> 1 }T
				T{ DOES1 ->   }T
				T{ CR1   -> 2 }T
				T{ DOES2 ->   }T
				T{ CR1   -> 3 }T
				T{ : WEIRD: CREATE DOES> 1 + DOES> 2 + ; -> }T
				T{ WEIRD: W1 -> }T
				T{ ' W1 >BODY -> HERE }T
				T{ W1 -> HERE 1 + }T
				T{ W1 -> HERE 2 + }T
			`,
		},
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
		// ELSE see IF
		{
			name: "EMIT",
			setup: `
						: OUTPUT-TEST
							." YOU SHOULD SEE THE STANDARD GRAPHIC CHARACTERS:" CR
							41 BL DO I EMIT LOOP CR
							61 41 DO I EMIT LOOP CR
							7F 61 DO I EMIT LOOP CR
							." YOU SHOULD SEE 0-9 SEPARATED BY A SPACE:" CR
							9 1+ 0 DO I . LOOP CR
							." YOU SHOULD SEE 0-9 (WITH NO SPACES):" CR
							[CHAR] 9 1+ [CHAR] 0 DO I 0 SPACES EMIT LOOP CR
							." YOU SHOULD SEE A-G SEPARATED BY A SPACE:" CR
							[CHAR] G 1+ [CHAR] A DO I EMIT SPACE LOOP CR
							." YOU SHOULD SEE 0-5 SEPARATED BY TWO SPACES:" CR
							5 1+ 0 DO I [CHAR] 0 + EMIT 2 SPACES LOOP CR
							." YOU SHOULD SEE TWO SEPARATE LINES:" CR
							S" LINE 1" TYPE CR S" LINE 2" TYPE CR
							." YOU SHOULD SEE THE NUMBER RANGES OF SIGNED AND UNSIGNED NUMBERS:" CR
							." SIGNED: " MIN-INT . MAX-INT . CR
							." UNSIGNED: " 0 U. MAX-UINT U. CR
						;
					`,
			code: `
						T{ OUTPUT-TEST -> }T
					`,
			expect: `YOU SHOULD SEE THE STANDARD GRAPHIC CHARACTERS:
 !"#$%&'()*+,-./0123456789:;<=>?@
ABCDEFGHIJKLMNOPQRSTUVWXYZ[\]^_` + "`" + `
abcdefghijklmnopqrstuvwxyz{|}~
YOU SHOULD SEE 0-9 SEPARATED BY A SPACE:
0 1 2 3 4 5 6 7 8 9 
YOU SHOULD SEE 0-9 (WITH NO SPACES):
0123456789
YOU SHOULD SEE A-G SEPARATED BY A SPACE:
A B C D E F G 
YOU SHOULD SEE 0-5 SEPARATED BY TWO SPACES:
0  1  2  3  4  5  
YOU SHOULD SEE TWO SEPARATE LINES:
LINE 1
LINE 2
YOU SHOULD SEE THE NUMBER RANGES OF SIGNED AND UNSIGNED NUMBERS:
SIGNED: -8000 7FFF 
UNSIGNED: 0 FFFF 
`,
		},
		// ENVIRONMENT? is not implemented
		{
			name: "EVALUATE",
			setup: `
				: GE1 S" 123" ; IMMEDIATE
				: GE2 S" 123 1+" ; IMMEDIATE
				: GE3 S" : GE4 345 ;" ;
				: GE5 EVALUATE ; IMMEDIATE
				T{ GE1 EVALUATE -> 123 }T \ TEST EVALUATE IN INTERPRET STATE
				T{ GE2 EVALUATE -> 124 }T
				T{ GE3 EVALUATE ->     }T
				T{ GE4          -> 345 }T

				T{ : GE6 GE1 GE5 ; -> }T \ TEST EVALUATE IN COMPILE STATE
				T{ GE6 -> 123 }T
				T{ : GE7 GE2 GE5 ; -> }T
				T{ GE7 -> 124 }T
			`,
		},
		// EXECUTE see ' [']
		// EXIT see UNLOOP
		{
			name: "FILL",
			setup: `
				T{ 20 CHARS BUFFER: FBUF -> }T
				T{ : SEEBUF FBUF C@ FBUF CHAR+ C@ FBUF CHAR+ CHAR+ C@ ; -> }T
			`,
			code: `
				T{ FBUF 0 20 FILL -> }T
				T{ SEEBUF -> 00 00 00 }T
				T{ FBUF 1 20 FILL -> }T
				T{ SEEBUF -> 20 00 00 }T

				T{ FBUF 3 20 FILL -> }T
				T{ SEEBUF -> 20 20 20 }T
			`,
		},
		{
			name: "FIND",
			setup: `
				HERE 3 C, CHAR G C, CHAR T C, CHAR 1 C, CONSTANT GT1STRING
				HERE 3 C, CHAR G C, CHAR T C, CHAR 2 C, CONSTANT GT2STRING
				HERE 3 ALLOT CONSTANT GT3STRING \ empty string size 3
				T{ GT1STRING FIND -> ' GT1 -1    }T
				T{ GT2STRING FIND -> ' GT2 1     }T
				T{ GT3STRING FIND -> GT3STRING 0 }T
			`,
		},
		// FM/MOD is not implemented
		// HERE see , ALLOT C,
		// HOLD is not implemented
		// I see LOOP LOOP+ J LEAVE UNLOOP
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
				T{ CREATE iw5 456 , IMMEDIATE -> }T
				T{ :NONAME iw5 [ @ iw3 ! ] ; DROP iw3 @ -> 456 }T
				T{ : iw6 CREATE , IMMEDIATE DOES> @ 1+ ; -> }T
				T{ 111 iw6 iw7 iw7 -> 112 }T
				T{ : iw8 iw7 LITERAL 1+ ; iw8 -> 113 }T
				T{ : iw9 CREATE , DOES> @ 2 + IMMEDIATE ; -> }T
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
		// KEY is not implemented
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
		// M* not implemented
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
		// MOD cannot test without double divide
		{
			name: "MOVE",
			setup: `
				T{ 20 BUFFER: FBUF -> }T
				T{ 20 FBUF ! 20 FBUF 1+ ! 20 FBUF 2 + ! -> }T
				T{ : SEEBUF FBUF @ FBUF 1+ @ FBUF 2 + @ ; -> }T
				T{ 20 BUFFER: SBUF -> }T
				T{ 12 SBUF ! 34 SBUF 1+ ! 56 SBUF 2 + ! -> }T
			`,
			code: `
				// The actual test suite is wrong for forths in which
				// the address units are greater than the character size.
				// This test is modified to work with that.
				T{ FBUF FBUF 3 CHARS MOVE -> }T \ BIZARRE SPECIAL CASE
				T{ SEEBUF -> 20 20 20 }T
				T{ SBUF FBUF 0 CHARS MOVE -> }T
				T{ SEEBUF -> 20 20 20 }T

				T{ SBUF FBUF 1 CHARS MOVE -> }T
				T{ SEEBUF -> 12 20 20 }T

				T{ SBUF FBUF 3 MOVE -> }T
				T{ SEEBUF -> 12 34 56 }T

				T{ FBUF FBUF 1+ 2 MOVE -> }T
				T{ SEEBUF -> 12 12 34 }T

				T{ FBUF 1+ FBUF 2 MOVE -> }T
				T{ SEEBUF -> 12 34 34 }T
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
		// QUIT not implemented
		// R> see >R
		// R@ see >R
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
		// REPEAT see WHILE
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
		{
			name: "S\"", // S"
			setup: `
				T{ : GC4 S" XY" ; ->   }T
				T{ : GC5 S" A String"2DROP ; -> }T \ There is no space between the " and 2DROP
				\ This test is from an extended mechanic, S" can be interpreted
				T{ S" A String"2DROP -> }T    \ There is no space between the " and 2DROP
			`,
			code: `
				T{ GC4 SWAP DROP  -> 2 }T
				T{ GC4 DROP DUP C@ SWAP CHAR+ C@ -> 58 59 }T
				T{ GC5 -> }T
			`,
		},
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
		// SIGN not implemented
		// SM/REM not implemented
		// SOURCE not implemented
		{
			name:   "SPACE",
			code:   "T{ SPACE -> }T",
			expect: " ",
		},
		{
			name: "SPACES",
			code: `
				T{ 0 SPACES    -> }T
				T{ 3 SPACES    -> }T
				T{ -100 SPACES -> }T
				T{ -1 SPACES   -> }T
			`,
			expect: "   ",
		},
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
		// THEN see IF
		// TYPE see EMIT
		// U. see EMIT
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
		// UM*
		// UM/MOD not implemented
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
		{
			name:  "VARIABLE",
			setup: "T{ VARIABLE V1 ->     }T",
			code: `
				T{    123 V1 ! ->     }T
				T{        V1 @ -> 123 }T
			`,
		},
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
		{
			name: "WORD",
			setup: `
				: GS3 WORD COUNT SWAP C@ ;
				T{ BL GS3 HELLO -> 5 CHAR H }T
				T{ CHAR " GS3 GOODBYE" -> 7 CHAR G }T
				T{ BL GS3 
				DROP -> 0 }T \ Blank lines return zero-length strings
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
			name:  "[",
			setup: "T{ : GC3 [ GC1 ] LITERAL ; -> }T",
			code:  "T{ GC3 -> 58 }T",
		},
		{
			name: "[']",
			code: "T{ POSTPONE GT2 EXECUTE -> 123 }T", // postpone because it's inside "main"
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
		// ] see [
	}
	runTests(t, tests)
}

func TestCoreExtensionSuite(t *testing.T) {
	tests := []suiteTest{
		{
			name: ".(",
			setup: `
				T{ : pb1 .( You should see 2345: ).( 2345); -> }T
			`,
			code: `
				T{ pb1 -> }T
			`,
			expect: "You should see 2345: 2345",
		},
		// .R not implemented
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
			name: "0>",
			code: `
				T{       0 0> -> <FALSE> }T
				T{      -1 0> -> <FALSE> }T
				T{ MIN-INT 0> -> <FALSE> }T
				T{       1 0> -> <TRUE>  }T
				T{ MAX-INT 0> -> <TRUE>  }T
			`,
		},
		// 2>R does not have any test cases
		// 2R> does not have any test cases
		// 2R@ does not have any test cases
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
		{
			name: "BUFFER:",
			setup: `
				DECIMAL
				T{ 127 CHARS BUFFER: TBUF1 -> }T
				T{ 127 CHARS BUFFER: TBUF2 -> }T \ Buffer is aligned
				T{ TBUF1 ALIGNED -> TBUF1 }T \ Buffers do not overlap
				\ T{ TBUF2 TBUF1 - ABS 127 CHARS < -> <FALSE> }T \ Buffer can be written to
				1 CHARS CONSTANT /CHAR
				: TFULL? ( c-addr n char -- flag )
				   TRUE 2SWAP CHARS OVER + SWAP ?DO
				     OVER I C@ = AND
				   /CHAR +LOOP NIP
				;
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
			name: "C\"",
			setup: `
				T{ : cq1 C" 123" ; -> }T
				T{ : cq2 C" " ;    -> }T
				T{ cq1 COUNT EVALUATE -> 123 }T
				T{ cq2 COUNT EVALUATE ->     }T
				\ This test is nonstandard, C" can be interpreted
				T{ C" A String"DROP -> }T    \ There is no space between the " and 2DROP
			`,
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
		{
			name: "COMPILE,",
			setup: `
				:NONAME DUP + ; CONSTANT dup+
				T{ : q dup+ COMPILE, ; -> }T
				T{ : as [ q ] ; -> }T
			`,
			code: `
				T{ 123 as -> 246 }T
			`,
		},
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
		// ENDCASE see CASE
		// ENDOF see case
		// ERASE does not have tests
		{
			name: "FALSE",
			code: `
				T{ FALSE -> 0 }T
				T{ FALSE -> <FALSE> }T
			`,
		},
		// HEX see BASE
		// HOLDS not implemented
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
		// MARKER not implemented
		{
			name: "NIP",
			code: `
				T{ 1 2 NIP -> 2 }T
			`,
		},
		// OF see CASE
		// PAD not implemented
		// PARSE not implemented
		// PARSE-NAME not implemented
		{
			name: "PICK",
			code: `
				T{ 123 456 789 2 PICK -> 123 456 789 123 }T
			`,
		},
		// REFILL not implemented
		// RESTORE-INPUT not implemented
		{
			name: "ROLL",
			code: `
				T{ 1 2 3 2 ROLL -> 2 3 1 }T
				T{ 1 2 1 ROLL -> 2 1 }T
				T{ 4 5 6 0 ROLL -> 4 5 6 }T
			`,
		},
		// S\" not implemented
		// SAVE-INPUT not implemented
		// SOURCE-ID not implemented
		// TO not implemented
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
		// U.R not implemented
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
		// UNUSED not implemented
		{
			name: "VALUE",
			setup: `
				T{  111 VALUE v1 -> }T
				T{ -999 VALUE v2 -> }T
				T{ : vd1 v1 ;    -> }T
				T{ : vd2 TO v2 ; -> }T

				T{ 123 VALUE v3 ->     }T
				T{ v3           -> 123 }T
				T{ 456 TO v3    ->     }T
				T{ v3           -> 456 }T
			`,
			code: `
				T{ v1 ->  111 }T
				T{ v2 -> -999 }T
				T{ 222 TO v1 -> }T
				T{ v1 -> 222 }T
				T{ vd1 -> 222 }T

				T{ v2 -> -999 }T
				T{ -333 vd2 -> }T
				T{ v2 -> -333 }T
				T{ v1 ->  222 }T
			`,
		},
		// WITHIN does not have test cases
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
			name:  "\\", // \
			setup: "T{ : COMMENT POSTPONE \\ ; IMMEDIATE -> }T",
			code: `
				T{ COMMENT \-postpone-fail
				-> }T
				\ \-inline-fail
			`,
		},
	}
	runTests(t, tests)
}

func TestDoubleSuite(t *testing.T) {
	tests := []suiteTest{
		{
			name: "2CONSTANT",
			setup: `
				T{ 1 2 2CONSTANT 2c1 -> }T
				T{ : cd1 2c1 ; -> }T

				T{ : cd2 2CONSTANT ; -> }T
				T{ -1 -2 cd2 2c2 -> }T

				T{ 4 5 2CONSTANT 2c3 IMMEDIATE 2c3 -> 4 5 }T
				T{ : cd6 2c3 2LITERAL ; -> }T
			`,
			code: `
				T{ 2c1 -> 1 2 }T
				T{ cd1 -> 1 2 }T

				
				T{ 2c2 -> -1 -2 }T

				T{ cd6 -> 4 5 }T
			`,
		},
		{
			name: "2LITERAL",
			setup: `
				T{ : cd1 [ MAX-2INT ] 2LITERAL ; -> }T
				
				T{ 2VARIABLE 2v4 IMMEDIATE 5 6 2v4 2! -> }T
				T{ : cd7 2v4 [ 2@ ] 2LITERAL ; ->  }T
				T{ : cd8 [ 6 7 ] 2v4 [ 2! ] ; 2v4 2@ -> 6 7 }T
			`,
			code: `
				T{ cd1 -> MAX-2INT }T
				T{ cd7 -> 5 6 }T
			`,
		},
		{
			name: "2VARIABLE",
			setup: `
				T{ 2VARIABLE 2v1 -> }T

				T{ : cd2 2VARIABLE ; -> }T
				T{ cd2 2v2 -> }T
				
				T{ : cd3 2v2 2! ; -> }T
				
				T{ 2VARIABLE 2v3 IMMEDIATE 5 6 2v3 2! -> }T
				T{ 2v3 2@ -> 5 6 }T
			`,
			code: `
				T{ 0. 2v1 2! ->    }T
				T{    2v1 2@ -> 0. }T
				T{ -1 -2 2v1 2! ->       }T
				T{       2v1 2@ -> -1 -2 }T

				T{ -2 -1 cd3 -> }T
				T{ 2v2 2@ -> -2 -1 }T
			`,
		},
		{
			name: "D+",
			code: `
				T{  0.  5. D+ ->  5. }T                         \ small integers
				T{ -5.  0. D+ -> -5. }T
				T{  1.  2. D+ ->  3. }T
				T{  1. -2. D+ -> -1. }T
				T{ -1.  2. D+ ->  1. }T
				T{ -1. -2. D+ -> -3. }T
				T{ -1.  1. D+ ->  0. }T
				T{  0  0  0  5 D+ ->  0  5 }T                  \ mid range integers
				T{ -1  5  0  0 D+ -> -1  5 }T
				T{  0  0  0 -5 D+ ->  0 -5 }T
				T{  0 -5 -1  0 D+ -> -1 -5 }T
				T{  0  1  0  2 D+ ->  0  3 }T
				T{ -1  1  0 -2 D+ -> -1 -1 }T
				T{  0 -1  0  2 D+ ->  0  1 }T
				T{  0 -1 -1 -2 D+ -> -1 -3 }T
				T{ -1 -1  0  1 D+ -> -1  0 }T

				T{ MIN-INT 0 2DUP D+ -> 0 1 }T
				T{ MIN-INT S>D MIN-INT 0 D+ -> 0 0 }T

				T{  HI-2INT       1. D+ -> 0 HI-INT 1+ }T    \ large double integers
				T{  HI-2INT     2DUP D+ -> 1S 1- MAX-INT }T
				T{ MAX-2INT MIN-2INT D+ -> -1. }T
				T{ MAX-2INT  LO-2INT D+ -> HI-2INT }T
				T{  LO-2INT     2DUP D+ -> MIN-2INT }T
				T{  HI-2INT MIN-2INT D+ 1. D+ -> LO-2INT }T
			`,
		},
		{
			name: "D-",
			code: `
				T{  0.  5. D- -> -5. }T              \ small integers
				T{  5.  0. D- ->  5. }T
				T{  0. -5. D- ->  5. }T
				T{  1.  2. D- -> -1. }T
				T{  1. -2. D- ->  3. }T
				T{ -1.  2. D- -> -3. }T
				T{ -1. -2. D- ->  1. }T
				T{ -1. -1. D- ->  0. }T
				T{  0  0  0  5 D- ->  0 -5 }T       \ mid-range integers
				T{ -1  5  0  0 D- -> -1  5 }T
				T{  0  0 -1 -5 D- ->  1  4 }T
				T{  0 -5  0  0 D- ->  0 -5 }T
				T{ -1  1  0  2 D- -> -1 -1 }T
				T{  0  1 -1 -2 D- ->  1  2 }T
				T{  0 -1  0  2 D- ->  0 -3 }T
				T{  0 -1  0 -2 D- ->  0  1 }T
				T{  0  0  0  1 D- ->  0 -1 }T
				T{ MIN-INT 0 2DUP D- -> 0. }T
				// T{ MIN-INT S>D MAX-INT 0D- -> 1 1s }T
				T{ MAX-2INT max-2INT D- -> 0. }T    \ large integers
				T{ MIN-2INT min-2INT D- -> 0. }T
				T{ MAX-2INT  hi-2INT D- -> lo-2INT DNEGATE }T
				T{  HI-2INT  lo-2INT D- -> max-2INT }T
				// T{  LO-2INT  hi-2INT D- -> min-2INT 1. D+ }T
				T{ MIN-2INT min-2INT D- -> 0. }T
				T{ MIN-2INT  lo-2INT D- -> lo-2INT }T \ TODO fixme
			`,
		},
		// D. not implemented
		// D.R not implemented
		{
			name: "D0<",
			code: `
				T{                0. D0< -> <FALSE> }T
				T{                1. D0< -> <FALSE> }T
				T{  MIN-INT        0 D0< -> <FALSE> }T
				T{        0  MAX-INT D0< -> <FALSE> }T
				T{          MAX-2INT D0< -> <FALSE> }T
				T{               -1. D0< -> <TRUE>  }T
				T{          MIN-2INT D0< -> <TRUE>  }T
			`,
		},
		{
			name: "D0=",
			code: `
				T{               1. D0= -> <FALSE> }T
				T{ MIN-INT        0 D0= -> <FALSE> }T
				T{         MAX-2INT D0= -> <FALSE> }T
				T{      -1  MAX-INT D0= -> <FALSE> }T
				T{               0. D0= -> <TRUE>  }T
				T{              -1. D0= -> <FALSE> }T
				T{       0  MIN-INT D0= -> <FALSE> }T
			`,
		},
		// D2* not implemented
		// D2/ not implemented
		{
			name: "D<",
			code: `
				T{       0.       1. D< -> <TRUE>  }T
				T{       0.       0. D< -> <FALSE> }T
				T{       1.       0. D< -> <FALSE> }T
				T{      -1.       1. D< -> <TRUE>  }T
				T{      -1.       0. D< -> <TRUE>  }T
				T{      -2.      -1. D< -> <TRUE>  }T
				T{      -1.      -2. D< -> <FALSE> }T
				T{      -1. MAX-2INT D< -> <TRUE>  }T
				T{ MIN-2INT MAX-2INT D< -> <TRUE>  }T
				T{ MAX-2INT      -1. D< -> <FALSE> }T
				T{ MAX-2INT MIN-2INT D< -> <FALSE> }T
				// T{ MAX-2INT 2DUP -1. D+ D< -> <FALSE> }T
				// T{ MIN-2INT 2DUP  1. D+ D< -> <TRUE>  }T
			`,
		},
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
		{
			name: "D>S",
			code: `
				T{    1234  0 D>S ->  1234   }T
				T{   -1234 -1 D>S -> -1234   }T
				T{ MAX-INT  0 D>S -> MAX-INT }T
				T{ MIN-INT -1 D>S -> MIN-INT }T
			`,
		},
		{
			name: "DABS",
			code: `
				T{       1. DABS -> 1.       }T
				T{      -1. DABS -> 1.       }T
				T{ MAX-2INT DABS -> MAX-2INT }T
				// T{ MIN-2INT 1. D+ DABS -> MAX-2INT }T
			`,
		},
		{
			name: "DMAX",
			code: `
				T{       1.       2. DMAX ->  2.      }T
				T{       1.       0. DMAX ->  1.      }T
				T{       1.      -1. DMAX ->  1.      }T
				T{       1.       1. DMAX ->  1.      }T
				T{       0.       1. DMAX ->  1.      }T
				T{       0.      -1. DMAX ->  0.      }T
				T{      -1.       1. DMAX ->  1.      }T
				T{      -1.      -2. DMAX -> -1.      }T
				T{ MAX-2INT  HI-2INT DMAX -> MAX-2INT }T
				T{ MAX-2INT MIN-2INT DMAX -> MAX-2INT }T
				T{ MIN-2INT MAX-2INT DMAX -> MAX-2INT }T
				T{ MIN-2INT  LO-2INT DMAX -> LO-2INT  }T
				T{ MAX-2INT       1. DMAX -> MAX-2INT }T
				T{ MAX-2INT      -1. DMAX -> MAX-2INT }T
				T{ MIN-2INT       1. DMAX ->  1.      }T
				T{ MIN-2INT      -1. DMAX -> -1.      }T
			`,
		},
		{
			name: "DMIN",
			code: `
				T{       1.       2. DMIN ->  1.      }T
				T{       1.       0. DMIN ->  0.      }T
				T{       1.      -1. DMIN -> -1.      }T
				T{       1.       1. DMIN ->  1.      }T
				T{       0.       1. DMIN ->  0.      }T
				T{       0.      -1. DMIN -> -1.      }T
				T{      -1.       1. DMIN -> -1.      }T
				T{      -1.      -2. DMIN -> -2.      }T
				T{ MAX-2INT  HI-2INT DMIN -> HI-2INT  }T
				T{ MAX-2INT MIN-2INT DMIN -> MIN-2INT }T
				T{ MIN-2INT MAX-2INT DMIN -> MIN-2INT }T
				T{ MIN-2INT  LO-2INT DMIN -> MIN-2INT }T
				T{ MAX-2INT       1. DMIN ->  1.      }T
				T{ MAX-2INT      -1. DMIN -> -1.      }T
				T{ MIN-2INT       1. DMIN -> MIN-2INT }T
				T{ MIN-2INT      -1. DMIN -> MIN-2INT }T
			`,
		},
		{
			name: "DNEGATE",
			code: `
				T{   0. DNEGATE ->  0. }T
				T{   1. DNEGATE -> -1. }T
				T{  -1. DNEGATE ->  1. }T
				T{ max-2int DNEGATE -> min-2int SWAP 1+ SWAP }T
				T{ min-2int SWAP 1+ SWAP DNEGATE -> max-2int }T
			`,
		},
		// M*/ not implemented
		{
			name: "M+",
			code: `
				T{ HI-2INT   1 M+ -> HI-2INT   1. D+ }T
				T{ MAX-2INT -1 M+ -> MAX-2INT -1. D+ }T
				T{ MIN-2INT  1 M+ -> MIN-2INT  1. D+ }T
				T{ LO-2INT  -1 M+ -> LO-2INT  -1. D+ }T
			`,
		},
	}
	runTests(t, tests)
}

func TestDoubleExtensionSuite(t *testing.T) {
	tests := []suiteTest{
		{
			name: "2ROT",
			code: `
				T{       1.       2. 3. 2ROT ->       2. 3.       1. }T
				T{ MAX-2INT MIN-2INT 1. 2ROT -> MIN-2INT 1. MAX-2INT }T
			`,
		},
		{
			name: "2VALUE",
			setup: `
				T{ 1 2 2VALUE t2val -> }T
				: sett2val t2val 2SWAP TO t2val ;

				T{ 1 2 2VALUE t2val2 -> }T
				T{ t2val2 -> 1 2 }T
				T{ 3 4 to t2val2 -> }T
				T{ t2val2 -> 3 4 }T
			`,
			code: `
				T{ t2val -> 1 2 }T
				T{ 3 4 TO t2val -> }T
				T{ t2val -> 3 4 }T
				T{ 5 6 sett2val t2val -> 3 4 5 6 }T
			`,
		},
		{
			name: "DU<",
			code: `
				T{       1.       1. DU< -> <FALSE> }T
				T{       1.      -1. DU< -> <TRUE>  }T
				T{      -1.       1. DU< -> <FALSE> }T
				T{      -1.      -2. DU< -> <FALSE> }T
				T{ MAX-2INT  HI-2INT DU< -> <FALSE> }T
				T{  HI-2INT MAX-2INT DU< -> <TRUE>  }T
				T{ MAX-2INT MIN-2INT DU< -> <TRUE>  }T
				T{ MIN-2INT MAX-2INT DU< -> <FALSE> }T
				T{ MIN-2INT  LO-2INT DU< -> <TRUE>  }T
			`,
		},
		{
			name: "DU>", // not standard
			code: `
				T{       1.       1. DU> -> <FALSE> }T
				T{       1.      -1. DU> -> <FALSE>  }T
				T{      -1.       1. DU> -> <TRUE> }T
				T{      -1.      -2. DU> -> <TRUE> }T
				T{ MAX-2INT  HI-2INT DU> -> <TRUE> }T
				T{  HI-2INT MAX-2INT DU> -> <FALSE>  }T
				T{ MAX-2INT MIN-2INT DU> -> <FALSE>  }T
				T{ MIN-2INT MAX-2INT DU> -> <TRUE> }T
				T{ MIN-2INT  LO-2INT DU> -> <FALSE>  }T
			`,
		},
	}
	runTests(t, tests)
}

func TestMemorySuite(t *testing.T) {
	tests := []suiteTest{
		{
			name: "ALLOCATE",
			setup: `
				T{ 50 CELLS ALLOCATE SWAP addr ! -> 0 }T
				T{ addr @ ALIGNED -> addr @ }T    \ Test address is aligned
				T{ HERE -> datsp @ }T              \ Check data space pointer is unaffected
				addr @ 50 write-cell-mem
				addr @ 50 check-cell-mem         \ Check we can access the heap
				T{ addr @ FREE -> 0 }T

				T{ 99 ALLOCATE SWAP addr ! -> 0 }T
				T{ addr @ ALIGNED -> addr @ }T    \ Test address is aligned
				T{ addr @ FREE -> 0 }T
				T{ HERE -> datsp @ }T              \ Data space pointer unaffected by FREE
				T{ -1 ALLOCATE SWAP DROP 0= -> <TRUE> }T    \ Memory allocation works with max size
			`,
		},
		{
			name: "RESIZE",
			setup: `
				T{ 50 CHARS ALLOCATE SWAP addr ! -> 0 }T
				addr @ 50 write-char-mem addr @ 50 check-char-mem
				\ Resize smaller does not change content.
				T{ addr @ 28 CHARS RESIZE SWAP addr ! -> 0 }T
				addr @ 28 check-char-mem

				\ Resize larger does not change original content.
				T{ addr @ 100 CHARS RESIZE SWAP addr ! -> 0 }T
				addr @ 28 check-char-mem

				\ We can resize to the maximum size
				T{ addr @ -1 RESIZE 0= -> addr @ <TRUE> }T

				T{ addr @ FREE -> 0 }T
				T{ HERE -> datsp @ }T    \ Data space pointer is unaffected
			`,
			code: ``,
		},
		// FREE does not have any tests
	}
	runTests(t, tests)
}

func runTests(t *testing.T, tests []suiteTest) {
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
			expect := tt.expect + "DONE"
			expect = strings.ReplaceAll(expect, "\n", "\r\n")
			runOutputTest(code, expect, t, &r)
		})
	}
}

func createTest(setup, code string) string {
	return fmt.Sprintf("%s %s : MAIN %s .\" DONE\" ESP.DONE ; ", suite, setup, code)
}
