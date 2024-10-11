
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
: T{ ( -- ) ;

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

: <TRUE> TRUE ;
: <FALSE> FALSE ;
: MAX-UINT 0xFFFF ;
: MAX-INT 0x7FFF ;
: MIN-UINT 0 ;
: MIN-INT 0x8000 ;
: MID-UINT 0x7FFF ;
: MID-UINT+1 0x8000 ;
: 0S 0 ;
: 1S 0xFFFF ;

HEX \ the test suite runs in hex mode

T{ : GC1 'X' ; -> }T \ should be : GC1 [CHAR] X ;
T{ : GC2 'H' ; -> }T \ should be : GC2 [CHAR] HELLO ;
T{ : GC3 [ GC1 ] LITERAL ; }T
: GN2 ( -- 16 10 )
    BASE @ >R HEX BASE @ DECIMAL BASE @ R> BASE ! ;

T{ : GT1 123 ; -> }T
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
VARIABLE nn1
VARIABLE nn2
T{ :NONAME 1234 ; nn1 ! -> }T
T{ :NONAME 9876 ; nn2 ! -> }T

\ from the DEFER test
T{ DEFER defer2 -> }T

\ from the DEFER@ test
T{ DEFER defer4 -> }T

\ from the DEFER! test
T{ DEFER defer3 -> }T

\ from the IS test
T{ DEFER defer5 -> }T
T{ : is-defer5 IS defer5 ; -> }T

RESET-TEST
