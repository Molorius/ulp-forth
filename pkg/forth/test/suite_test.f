
VARIABLE TEST-COUNT
VARIABLE TEST-DEPTH
16 BUFFER: TEST-STACK \ might need to increase if tests need more space

\ run this to indicate that a test passes
: TEST-PASS ( -- ) ;

\ print the values on the stack
: printstack
    1 DEPTH 1- ?DO
        I 1-   \ get the offset
        PICK . \ pick the value and print
    -1 +LOOP
;

\ print the values on the TEST-STACK buffer
: printteststack
    TEST-DEPTH @ 0 ?DO
        I TEST-STACK + \ get the address
        @ . \ read the address and print
    LOOP
;

: TEST-FAIL ( -- )
    ."  test "
    TEST-COUNT @ U. \ print the test number
    ." got "
    printteststack
    ." expected "
    printstack
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

\ from the [CHAR] test, used in many
T{ : GC1 [CHAR] X     ; -> }T

\ from the ' test, used in many
T{ : GT1 123 ;   ->     }T

\ from the ['] test, used in many
T{ : GT2 ['] GT1 ; IMMEDIATE -> }T

\ ' EXIT CONSTANT 1ST
T{ VARIABLE 1ST -> }T \ not the same as the test suite

RESET-TEST
