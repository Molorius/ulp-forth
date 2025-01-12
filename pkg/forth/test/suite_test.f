\ Copyright 2024 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

VARIABLE TEST-COUNT
VARIABLE GOT-DEPTH
VARIABLE WANT-DEPTH
16 CONSTANT BUFFERDEPTH
BUFFERDEPTH BUFFER: GOT-STACK \ might need to increase if tests need more space
BUFFERDEPTH BUFFER: WANT-STACK

\ lets us use // as a comment
: // POSTPONE \ ; IMMEDIATE

\ run this to indicate that a test passes
: TEST-PASS ( -- ) ;

: BUFFERPRINT ( addr count -- )
    0 ?DO
        DUP I + @ .
    LOOP
    DROP
;

: TEST-FAIL ( -- )
    ."  test "
    TEST-COUNT @ U. \ print the test number
    ." got "
    GOT-STACK GOT-DEPTH @ BUFFERPRINT
    ." expected "
    WANT-STACK WANT-DEPTH @ BUFFERPRINT
;

\ begin a test
: T{ ( -- ) ;

: BUFFERFILL ( stackn ... stack0 bufaddr countaddr -- )
    DEPTH 2 - DUP ROT ! \ record the stack depth
    0<> IF \ if there are items on the stack
        DEPTH + 2 - \ get to the end
        BEGIN
            DUP ROT SWAP ( addr item addr )
            ! \ store the item into address
            1- \ ( addr-1 )
        DEPTH 1 = UNTIL \ loop until the stack only has addr+1
    THEN
    DROP \ drop address
;

: BUFFEREQUALS ( addr0 count0 addr1 count1 -- bool )
    2 PICK <> IF \ if count0 != count1
        FALSE EXIT \ return false
    THEN
    SWAP ( addr0 addr1 count )
    0 ?DO
        ( addr0 addr1 )
        OVER I + @ \ get value from addr0
        OVER I + @ \ get value from addr1
        <> IF \ if they're not equal
            2DROP UNLOOP FALSE EXIT \ return false
        THEN
    LOOP
    2DROP
    TRUE
;

\ save desired values to compare test result against
: -> ( -- )
    GOT-STACK GOT-DEPTH BUFFERFILL
;

: TEST ( stackn stackn-1 ... stack0 testdepth testn -- )
    GOT-STACK GOT-DEPTH @ WANT-STACK WANT-DEPTH @ BUFFEREQUALS
    0= IF
        TEST-FAIL VM.STACK.INIT EXIT \ fail the test, reset stack, exit
    THEN
;

: }T ( -- empties-the-stack)
    WANT-STACK WANT-DEPTH BUFFERFILL \ save the want results
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
MAX-INT 2/ CONSTANT HI-INT
MIN-INT 2/ CONSTANT LO-INT
0x7FFF CONSTANT MID-UINT
0x8000 CONSTANT MID-UINT+1
0      CONSTANT 0S
0xFFFF CONSTANT 1S
0x8000 CONSTANT MSB
0xFFFFFFFF. 2CONSTANT MAX-2UINT
0x7FFFFFFF. 2CONSTANT MAX-2INT
0.          2CONSTANT MIN-2UINT
0x80000000. 2CONSTANT MIN-2INT
MAX-2INT 2/ 2CONSTANT HI-2INT
MIN-2INT 2/ 2CONSTANT LO-2INT

HEX \ the test suite runs in hex mode

\ from the [CHAR] test, used in many
T{ : GC1 [CHAR] X     ; -> }T

\ from the ' test, used in many
T{ : GT1 123 ;   ->     }T

\ from the ['] test, used in many
T{ : GT2 ['] GT1 ; IMMEDIATE -> }T

\ ' EXIT CONSTANT 1ST
T{ VARIABLE 1ST -> }T \ not the same as the test suite

RESET-TEST
