
VARIABLE TEST-COUNT
VARIABLE TEST-DEPTH
16 BUFFER: TEST-STACK \ might need to increase if tests need more space

\ run this to indicate that a test passes
: TEST-PASS ( -- ) ; IMMEDIATE
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

\ from the IS test
T{ DEFER defer5 -> }T
T{ : is-defer5 IS defer5 ; -> }T
T{ ' * IS defer5 -> }T
T{ 2 3 defer5 -> 6 }T
T{ ' + is-defer5 -> }T
T{ 1 2 defer5 -> 3 }T

\ from the S" test
T{ : GC4 S" XY" ; ->   }T
T{ GC4 SWAP DROP  -> 2 }T
T{ GC4 DROP DUP C@ SWAP CHAR+ C@ -> 58 59 }T
: GC5 S" A String"2DROP ; \ There is no space between the " and 2DROP
T{ GC5 -> }T

RESET-TEST
