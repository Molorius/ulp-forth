
VARIABLE TEST-COUNT
VARIABLE TEST-DEPTH
16 BUFFER: TEST-STACK \ might need to increase if tests need more space

\ run this to indicate that a test passes
: TEST-PASS ( -- ) ;
\ TODO print better fail messages
: TEST-FAIL ( -- )
    BL EMIT \ print a space
    TEST-COUNT @ U. \ print the test number
    'F' EMIT \ print an 'F'
;

\ begin a test
: T{ ( -- ) ; IMMEDIATE

\ save desired values to compare test result against
: -> ( -- )
    DEPTH TEST-DEPTH ! \ record the stack depth
    DEPTH 0<> IF \ if there are items on the stack
        TEST-STACK \ get the buffer address
        BEGIN
            DUP ROT SWAP ( addr item addr )
            ! \ store the item into address
            1+ \ ( addr+1 )
        DEPTH 1 = UNTIL \ loop until the stack only has addr+1
        DROP \ drop address
    THEN
;

: TEST ( stackn stackn-1 ... stack0 testdepth testn -- )
    DEPTH \ get the depth
    TEST-DEPTH @ \ get the test depth
    <> IF \ if they're not equal
        TEST-FAIL VM.STACK.INIT EXIT \ fail the test, reset stack, exit
    THEN

    DEPTH 0<> IF \ if there are items on the stack
        TEST-STACK \ get the buffer address
        BEGIN
            SWAP OVER ( addr got addr )
            @ ( addr got expect )
            <> IF \ if they're not equal
                TEST-FAIL VM.STACK.INIT EXIT \ fail!
            THEN
            1+ ( addr+1 )
        DEPTH 1 = UNTIL \ loop until stack only has addr+1
        DROP \ drop address
    THEN
;

: }T ( -- empties-the-stack)
    TEST \ run the test
    1 TEST-COUNT +! \ increment test number
;

\ reset the test count
: RESET-TEST
    0 TEST-COUNT !
;

TRUE   CONSTANT <TRUE>
FALSE  CONSTANT <FALSE>
0xFFFF CONSTANT MAX-UINT
0x7FFF CONSTANT MAX-INT
0      CONSTANT MIN-UINT
0x8000 CONSTANT MIN-INT
0x7FFF CONSTANT MID-UINT
0x8000 CONSTANT MID-UINT+1
0      CONSTANT 0S
0xFFFF CONSTANT 1S
0x8000 CONSTANT MSB

HEX \ the test suite runs in hex mode

\ from the [CHAR] test
T{ : GC1 [CHAR] X     ; -> }T
T{ : GC2 [CHAR] HELLO ; -> }T
T{ GC1 -> 58 }T
T{ GC2 -> 48 }T

T{ : GC3 [ GC1 ] LITERAL ; -> }T
T{ : GN2 ( -- 16 10 )
    BASE @ >R HEX BASE @ DECIMAL BASE @ R> BASE ! ; -> }T

\ from the ' test
T{ : GT1 123 ;   ->     }T
T{ ' GT1 EXECUTE -> 123 }T

T{ : GT2 ['] GT1 ; IMMEDIATE -> }T

\ part of the CONSTANT test
T{ 123 CONSTANT X123 -> }T
T{ : EQU CONSTANT ; -> }T
T{ X123 EQU Y123 -> }T

\ part of the : test
T{ : NOP : POSTPONE ; ; -> }T
T{ NOP NOP1 NOP NOP2 -> }T
T{ : GDX 123 ; : GDX GDX 234 ; -> }T

\ from the :NONAME test
T{ VARIABLE nn1 -> }T
T{ VARIABLE nn2 -> }T
T{ :NONAME 1234 ; nn1 ! -> }T
T{ :NONAME 9876 ; nn2 ! -> }T

\ from the ACTION-OF test
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

\ from the BUFFER: test
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
HEX

\ from the CASE test
: cs1 CASE 1 OF 111 ENDOF
   2 OF 222 ENDOF
   3 OF 333 ENDOF
   >R 999 R>
   ENDCASE
;
T{ 1 cs1 -> 111 }T
T{ 2 cs1 -> 222 }T
T{ 3 cs1 -> 333 }T
T{ 4 cs1 -> 999 }T
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
T{ -1 1 cs2 ->  100 }T
T{ -1 2 cs2 ->  200 }T
T{ -1 3 cs2 -> -300 }T
T{ -2 1 cs2 ->  -99 }T
T{ -2 2 cs2 -> -199 }T
T{  0 2 cs2 ->  299 }T

\ from the [COMPILE] test
T{ : [c1] [COMPILE] DUP ; IMMEDIATE -> }T
T{ 123 [c1] -> 123 123 }T
T{ : [c2] [COMPILE] [c1] ; -> }T
T{ 234 [c2] -> 234 234 }T
T{ : [cif] [COMPILE] IF ; IMMEDIATE -> }T
T{ : [c3]  [cif] 111 ELSE 222 THEN ; -> }T
T{ -1 [c3] -> 111 }T
T{  0 [c3] -> 222 }T

\ from the DEFER test
T{ DEFER defer2 -> }T
T{ ' * ' defer2 DEFER! -> }T
T{   2 3 defer2 -> 6 }T
T{ ' + IS defer2 ->   }T
T{    1 2 defer2 -> 3 }T

\ from the DEFER@ test
T{ DEFER defer4 -> }T
T{ ' * ' defer4 DEFER! -> }T
T{ 2 3 defer4 -> 6 }T
T{ ' defer4 DEFER@ -> ' * }T
T{ ' + IS defer4 -> }T
T{ 1 2 defer4 -> 3 }T
T{ ' defer4 DEFER@ -> ' + }T

\ from the DEFER! test
T{ DEFER defer3 -> }T
T{ ' * ' defer3 DEFER! -> }T
T{ 2 3 defer3 -> 6 }T
T{ ' + ' defer3 DEFER! -> }T
T{ 1 2 defer3 -> 3 }T

\ modified from the FIND test
T{ BL WORD GT1 CONSTANT GT1STRING -> }T

\ from the IF test
T{ : GI1 IF 123 THEN ; -> }T
T{ : GI2 IF 123 ELSE 234 THEN ; -> }T
T{  0 GI1 ->     }T
T{  1 GI1 -> 123 }T
T{ -1 GI1 -> 123 }T
T{  0 GI2 -> 234 }T
T{  1 GI2 -> 123 }T
T{ -1 GI1 -> 123 }T
\ Multiple ELSEs in an IF statement
T{ : melse IF 1 ELSE 2 ELSE 3 ELSE 4 ELSE 5 THEN ; -> }T
T{ <FALSE> melse -> 2 4 }T
T{ <TRUE>  melse -> 1 3 5 }T

\ from the IMMEDIATE test
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

\ from the IS test
T{ DEFER defer5 -> }T
T{ : is-defer5 IS defer5 ; -> }T
T{ ' * IS defer5 -> }T
T{ 2 3 defer5 -> 6 }T
T{ ' + is-defer5 -> }T
T{ 1 2 defer5 -> 3 }T

\ for the \ test
T{ : COMMENT POSTPONE \ ; IMMEDIATE -> }T


\ from the LITERAL test
T{ : GT3 GT2 LITERAL ; -> }T
T{ GT3 -> ' GT1 }T

\ from the LOOP test
T{ : GD1 DO I LOOP ; -> }T
T{          4        1 GD1 ->  1 2 3   }T
T{          2       -1 GD1 -> -1 0 1   }T
T{ MID-UINT+1 MID-UINT GD1 -> MID-UINT }T

\ from the +LOOP test
T{ : GD2 DO I -1 +LOOP ; -> }T
T{        1          4 GD2 -> 4 3 2  1 }T
T{       -1          2 GD2 -> 2 1 0 -1 }T
T{ MID-UINT MID-UINT+1 GD2 -> MID-UINT+1 MID-UINT }T
\ VARIABLE gditerations
\ VARIABLE gdincrement
\ : gd7 ( limit start increment -- )
\    gdincrement !
\    0 gditerations !
\    DO
\      1 gditerations +!
\      I
\      gditerations @ 6 = IF LEAVE THEN
\      gdincrement @
\    +LOOP gditerations @
\ ;
\ T{    4  4  -1 gd7 ->  4                  1  }T
\ T{    1  4  -1 gd7 ->  4  3  2  1         4  }T
\ T{    4  1  -1 gd7 ->  1  0 -1 -2  -3  -4 6  }T
\ T{    4  1   0 gd7 ->  1  1  1  1   1   1 6  }T
\ T{    0  0   0 gd7 ->  0  0  0  0   0   0 6  }T
\ T{    1  4   0 gd7 ->  4  4  4  4   4   4 6  }T
\ T{    1  4   1 gd7 ->  4  5  6  7   8   9 6  }T
\ T{    4  1   1 gd7 ->  1  2  3            3  }T
\ T{    4  4   1 gd7 ->  4  5  6  7   8   9 6  }T
\ T{    2 -1  -1 gd7 -> -1 -2 -3 -4  -5  -6 6  }T
\ T{   -1  2  -1 gd7 ->  2  1  0 -1         4  }T
\ T{    2 -1   0 gd7 -> -1 -1 -1 -1  -1  -1 6  }T
\ T{   -1  2   0 gd7 ->  2  2  2  2   2   2 6  }T
\ T{   -1  2   1 gd7 ->  2  3  4  5   6   7 6  }T
\ T{    2 -1   1 gd7 -> -1 0 1              3  }T
\ T{  -20 30 -10 gd7 -> 30 20 10  0 -10 -20 6  }T
\ T{  -20 31 -10 gd7 -> 31 21 11  1  -9 -19 6  }T
\ T{  -20 29 -10 gd7 -> 29 19  9 -1 -11     5  }T
\ \ With large and small increments
\ MAX-UINT 8 RSHIFT 1+ CONSTANT ustep
\ ustep NEGATE CONSTANT -ustep
\ MAX-INT 7 RSHIFT 1+ CONSTANT step
\ step NEGATE CONSTANT -step
\ VARIABLE bump
\ T{  : gd8 bump ! DO 1+ bump @ +LOOP ; -> }T
\ T{  0 MAX-UINT 0 ustep gd8 -> 256 }T
\ T{  0 0 MAX-UINT -ustep gd8 -> 256 }T
\ T{  0 MAX-INT MIN-INT step gd8 -> 256 }T
\ T{  0 MIN-INT MAX-INT -step gd8 -> 256 }T

\ from the POSTPONE test
T{ : GT4 POSTPONE GT1 ; IMMEDIATE -> }T
T{ : GT5 GT4 ; -> }T
T{ GT5 -> 123 }T
T{ : GT6 345 ; IMMEDIATE -> }T
T{ : GT7 POSTPONE GT6 ; -> }T
T{ GT7 -> 345 }T

\ from the RECURSE test
T{ : GI6 ( N -- 0,1,..N ) 
     DUP IF DUP >R 1- RECURSE R> THEN ; -> }T
T{ 0 GI6 -> 0 }T
T{ 1 GI6 -> 0 1 }T
T{ 2 GI6 -> 0 1 2 }T
T{ 3 GI6 -> 0 1 2 3 }T
T{ 4 GI6 -> 0 1 2 3 4 }T
DECIMAL
T{ :NONAME ( n -- 0, 1, .., n ) 
     DUP IF DUP >R 1- RECURSE R> THEN 
   ; 
   CONSTANT rn1 -> }T
T{ 0 rn1 EXECUTE -> 0 }T
T{ 4 rn1 EXECUTE -> 0 1 2 3 4 }T
\ :NONAME ( n -- n1 )
\    1- DUP
\    CASE 0 OF EXIT ENDOF
\      1 OF 11 SWAP RECURSE ENDOF
\      2 OF 22 SWAP RECURSE ENDOF
\      3 OF 33 SWAP RECURSE ENDOF
\      DROP ABS RECURSE EXIT
\    ENDCASE
\ ; CONSTANT rn2
\ T{  1 rn2 EXECUTE -> 0 }T
\ T{  2 rn2 EXECUTE -> 11 0 }T
\ T{  4 rn2 EXECUTE -> 33 22 11 0 }T
\ T{ 25 rn2 EXECUTE -> 33 22 11 0 }T
HEX

\ from the S" test
T{ : GC4 S" XY" ; ->   }T
T{ GC4 SWAP DROP  -> 2 }T
T{ GC4 DROP DUP C@ SWAP CHAR+ C@ -> 58 59 }T
T{ : GC5 S" A String"2DROP ; -> }T \ There is no space between the " and 2DROP
T{ GC5 -> }T

\ from the STATE test
T{ : GT8 STATE @ ; IMMEDIATE -> }T
T{ GT8 -> 0 }T
T{ : GT9 GT8 LITERAL ; -> }T
T{ GT9 0= -> <FALSE> }T

\ from the UNTIL test
T{ : GI4 BEGIN DUP 1+ DUP 5 > UNTIL ; -> }T
T{ 3 GI4 -> 3 4 5 6 }T
T{ 5 GI4 -> 5 6 }T
T{ 6 GI4 -> 6 7 }T

\ from the VARIABLE test
T{ VARIABLE V1 ->     }T
T{    123 V1 ! ->     }T
T{        V1 @ -> 123 }T

\ from the WHILE test
T{ : GI3 BEGIN DUP 5 < WHILE DUP 1+ REPEAT ; -> }T
T{ 0 GI3 -> 0 1 2 3 4 5 }T
T{ 4 GI3 -> 4 5 }T
T{ 5 GI3 -> 5 }T
T{ 6 GI3 -> 6 }T
T{ : GI5 BEGIN DUP 2 > WHILE
      DUP 5 < WHILE DUP 1+ REPEAT
      123 ELSE 345 THEN ; -> }T
T{ 1 GI5 -> 1 345 }T
T{ 2 GI5 -> 2 345 }T
T{ 3 GI5 -> 3 4 5 123 }T
T{ 4 GI5 -> 4 5 123 }T
T{ 5 GI5 -> 5 123 }T

\ from the ( test
\ There is no space either side of the ).
T{ ( A comment)1234 -> 1234 }T
T{ : pc1 ( A comment)1234 ; pc1 -> 1234 }T

\ from the >R test
T{ : GR1 >R R> ; -> }T
T{ : GR2 >R R@ R> DROP ; -> }T
T{ 123 GR1 -> 123 }T
T{ 123 GR2 -> 123 }T
T{  1S GR1 ->  1S }T      ( Return stack holds cells )

\ ' EXIT CONSTANT 1ST
T{ VARIABLE 1ST -> }T \ not the same as the test suite

RESET-TEST
