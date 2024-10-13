
\ ['] parses the next name and compiles the execution token of that name. Immediate.
: ['] ( "<spaces>name" -- ) ' POSTPONE LITERAL ; IMMEDIATE

\ RECURSE compiles the most recently defined name, which usually means the current definition. 
\ Immediate.
: RECURSE ( -- ) LAST COMPILE, ; IMMEDIATE

: >BODY ( xt -- a-addr ) ; \ the address is the same as the xt in this kernel

: DEFER ( "<spaces>name" -- )
    CREATE \ create a new dictionary entry
    POSTPONE EXIT \ compile an EXIT here for now. can be changed by DEFER!.
    POSTPONE EXIT \ compile the actual EXIT.
;
: DEFER@ ( xt1 -- xt2 ) >BODY @ ;
: DEFER! ( xt2 xt1 -- ) >BODY ! ;

: NIP ( a b -- b ) SWAP DROP ;
: 1+ ( x -- x+1 ) 1 + ;
: 1- ( x -- x-1 ) 1 - ;
\ Negate x. -2 becomes 2, 3 becomes -3, etc.
: NEGATE ( x -- -x ) 0 SWAP - ;
\ Invert all bits. 0xFFFF becomes 0, 0xFFF0 becomes 0x000F, etc.
: INVERT ( x -- x^0xFFFF ) NEGATE 1- ;
: 2* 1 LSHIFT ;
: 2/
    DUP
    0x8000 AND \ isolate the most significant bit
    SWAP 1 RSHIFT \ right shift
    OR \ then put back the bit
;
: 2DROP DROP DROP ;
: CHARS
    DUP
    2/ \ divide by 2
    SWAP 1 AND \ isolate the lowest bit
    + \ then put back the pit
;

: U> SWAP U< ; \ greaterthan is just lessthan with operands swapped
: > 
    0x8000 - \ shift top of stack into unsigned space
    SWAP 0x8000 - \ shift next into unsigned space
    U< \ compare!
;
: < SWAP > ;

: IF
    BRANCH0 \ create a conditional branch
    DUP COMPILE, \ compile it
    >C \ and put it on the control flow stack
; IMMEDIATE

: ELSE
    BRANCH \ create a definite branch
    DUP COMPILE, \ compile it, keep copy on bottom of stack
    DEST DUP COMPILE, \ create a destination for previous branch, compile it
    C> \ take the previous branch off the control flow stack
    SWAP RESOLVE-BRANCH \ then resolve it
    >C \ put definite branch on control flow stack
; IMMEDIATE

: THEN
    C> \ take the branch off the control flow stack
    DEST DUP COMPILE, \ create a destination, compile it
    RESOLVE-BRANCH \ then resolve the branch
; IMMEDIATE

: BEGIN
    DEST DUP COMPILE, \ create a destination, compile it
    >C \ put it on the control flow stack
; IMMEDIATE

: UNTIL
    BRANCH0 DUP COMPILE, \ create a condition branch, compile it
    C> RESOLVE-BRANCH \ and resolve it
; IMMEDIATE

: AGAIN \ same logic as UNTIL but with a definite branch
    BRANCH DUP COMPILE,
    C> RESOLVE-BRANCH
; IMMEDIATE

: 0= IF FALSE ELSE TRUE THEN ;
: = - 0= ;
: 0<> 0= 0= ;
: <> = 0= ;
: 0< 0x7FFF U> ;
: 0> 1- 0x7FFF U< ;

: IS
    STATE @ IF
        POSTPONE ['] POSTPONE DEFER!
    ELSE
        ' DEFER!
    THEN
; IMMEDIATE

\ Parse the next word delimited by a space. Allocate n cells. Create
\ a definition for the word that places the address of the allocated
\ memory onto the stack.
: BUFFER: ( n -- )
    ALLOCATE DROP \ allocate n words, drop the superfluous "ok" indicator but keep address
    : \ parse the next input, create a word with that name
    POSTPONE LITERAL \ compile the allocated address literal
    POSTPONE ; \ end the definition
;

: VARIABLE 1 BUFFER: ;
: 2VARIABLE 2 BUFFER: ;
: +! DUP @ ROT + SWAP ! ;

: OVER 1 PICK ;
: 2OVER 3 PICK 3 PICK ;
: 2DUP OVER OVER ;
: 2>R POSTPONE SWAP POSTPONE >R POSTPONE >R ; IMMEDIATE
: 2R> POSTPONE R> POSTPONE R> POSTPONE SWAP ; IMMEDIATE
: UNLOOP POSTPONE >R POSTPONE >R POSTPONE 2DROP ; IMMEDIATE
: I POSTPONE R> POSTPONE DUP POSTPONE >R ; IMMEDIATE
: ?DUP DUP IF DUP THEN ;
: XOR ( a b -- c )
    \ [a ^ b] = [a|b] - [a&b]
    2DUP ( A B A B )
    OR ( A B [A|B] )
    SWAP ROT ( [A|B] B A )
    AND ( [A|B] [A&B] )
    - ( [A^B] )
;
: ABS DUP 0< IF NEGATE THEN ;

: --CONSTANT
    STATE @ IF \ if we're compiling
        POSTPONE LITERAL \ then compile the value on the stack
    THEN \ otherwise leave on the stack
;

: CONSTANT
    :
    POSTPONE LITERAL
    POSTPONE --CONSTANT
    POSTPONE ;
    IMMEDIATE
;

: ACTION-OF
    STATE @ IF
        POSTPONE ['] POSTPONE DEFER@
    ELSE
        ' DEFER@
    THEN
; IMMEDIATE

: HEX 16 BASE ! ;
: DECIMAL 10 BASE ! ;
: CELL+ 1+ ;
: CELLS ;
: [COMPILE]
    ' \ get the execution token of the next input
    COMPILE, \ and compile it!
; IMMEDIATE
: COUNT ( c-addr -- c-addr+1 n ) DUP CHAR+ SWAP C@ ;
: S" '"' WORD COUNT SWAP POSTPONE LITERAL POSTPONE LITERAL ; IMMEDIATE

\ : ?DO
\     POSTPONE 2DUP \ compile 2dup
\     POSTPONE 2>R \ compile 2>r
\     POSTPONE <> \ compile <>
\     BRANCH0 DUP COMPILE, \ compile a conditional branch
\     DEST DUP COMPILE, \ compile a destination
\     >C >C \ push destination then branch on control stack
\ ; IMMEDIATE
