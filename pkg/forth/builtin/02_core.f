
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

: WHILE
    BRANCH0 \ create a conditional branch
    DUP COMPILE, \ compile it
    C> ( branch0 dest )
    SWAP ( dest branch0 )
    >C >C ( C: branch0 dest )
; IMMEDIATE

: REPEAT
    BRANCH \ create a branch
    DUP COMPILE, \ compile it ( repeat-branch )
    C> RESOLVE-BRANCH \ and resolve it
    C> \ get the while-branch0
    DEST DUP COMPILE, \ compile a destination for the while-branch0
    RESOLVE-BRANCH \ resolve while-branch0
; IMMEDIATE

: 0= IF FALSE EXIT THEN TRUE ;
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
: +! DUP @ ROT + SWAP ! ;

: OVER 1 PICK ;
: 2OVER 3 PICK 3 PICK ;
: 2DUP OVER OVER ;
: 2>R POSTPONE SWAP POSTPONE >R POSTPONE >R ; IMMEDIATE
: 2R> POSTPONE R> POSTPONE R> POSTPONE SWAP ; IMMEDIATE
: R@ 0 POSTPONE LITERAL POSTPONE RPICK ; IMMEDIATE
: 2R@ 1 POSTPONE LITERAL POSTPONE RPICK POSTPONE R@ ; IMMEDIATE
: UNLOOP POSTPONE >R POSTPONE >R POSTPONE 2DROP ; IMMEDIATE
: I 0 POSTPONE LITERAL POSTPONE RPICK ; IMMEDIATE
: J 2 POSTPONE LITERAL POSTPONE RPICK ; IMMEDIATE
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
: MAX 2DUP < IF SWAP THEN DROP ;
: MIN 2DUP > IF SWAP THEN DROP ;
: S>D  ( n -- d) \ convert the signed number to double-cell
    DUP \ duplicate number
    0x7FFF U> \ if greater then the largest int, set to all 1s
;
: U>D 0 ( u -- d ) ; \ convert the unsigned number to double-cell
: TUCK SWAP OVER ;
: WITHIN ( test low high -- flag ) OVER - >R - R> U< ;

: CONSTANT
    :
    POSTPONE LITERAL
    POSTPONE ;
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
: STRING" '"' WORD ; \ read a string and put it on the stack
: C" STRING" POSTPONE LITERAL ; IMMEDIATE
: S" STRING" COUNT SWAP POSTPONE LITERAL POSTPONE LITERAL ; IMMEDIATE
: [CHAR]
    BL WORD \ get the next word
    COUNT DROP \ get the address of the first letter
    C@ \ read the letter
    POSTPONE LITERAL \ then compile it!
; IMMEDIATE
: 2@ DUP CELL+ @ SWAP @ ;
: 2! SWAP OVER ! CELL+ ! ;
: 2SWAP ROT >R ROT R> ;


: CASE ( C: -- case-sys )
    0 >C \ put a 0 on the control stack as a marker
; IMMEDIATE

: OF
    POSTPONE OVER \ copy the lower number
    POSTPONE = \ check if upper == lower
    POSTPONE IF \ create a branch
    POSTPONE DROP
; IMMEDIATE

: ENDOF
    POSTPONE ELSE \ resolve the OF and create a new branch
; IMMEDIATE

: ENDCASE
    POSTPONE DROP
    BEGIN
        C> ?DUP \ get the top of control flow stack, copy if not 0 (a branch)
    WHILE
        >C \ not a 0 so put back on control flow stack
        POSTPONE THEN \ and resolve the branch
    REPEAT
; IMMEDIATE

: DO ( C: -- dest )
    POSTPONE 2>R \ compile 2>r
    DEST DUP COMPILE, \ create and compile a destination
    >C \ and push destination on the control flow stack
    0 >D \ push 0 onto the DO stack
; IMMEDIATE

: ?DO
    POSTPONE 2DUP \ dupe the inputs
    POSTPONE 2>R \ put one copy on return stack
    POSTPONE <> \ check if the others are equal
    0 >D \ push 0 on to the DO stack
    BRANCH0 DUP COMPILE, \ compile a conditional branch
    >D \ and push onto the DO stack
    DEST DUP COMPILE, \ compile a destination
    >C \ and push onto the control flow stack
; IMMEDIATE

: UNLOOP
    POSTPONE R> POSTPONE R> POSTPONE 2DROP
; IMMEDIATE

: +LOOP
    POSTPONE LOOPCHECK \ compile the check
    BRANCH0 DUP COMPILE, \ create and compile a conditional branch
    C> RESOLVE-BRANCH \ and resolve it
    BEGIN
        D> ?DUP \ get the top of DO stack, copy if not 0
    WHILE
        \ the item is a branch!
        DEST DUP COMPILE, \ create and compile a destination
        RESOLVE-BRANCH \ and resolve the branch
    REPEAT
    POSTPONE UNLOOP \ remove loop items
; IMMEDIATE

: LOOP
    1 POSTPONE LITERAL \ compile the number 1
    POSTPONE +LOOP \ and +LOOP
; IMMEDIATE

: LEAVE
    BRANCH DUP COMPILE, \ create and compile a branch
    >D \ also push it onto the DO stack
; IMMEDIATE

: F/MOD ( numerator denominator -- remainder quotient )
    \ we have the unsigned version U/MOD as a primitive
    DUP 0< DUP >R \ check if denominator is negative and store on return stack
    IF NEGATE THEN \ if negative then negate
    SWAP
    DUP 0< DUP >R \ check if numerator is negative and store on return stack
    IF NEGATE THEN \ if negative then negate
    SWAP
    U/MOD \ unsigned divide!
    \ R> R> 2DROP
    R> R@ <> IF NEGATE THEN \ if num and den were different signs then negate the quotient
    SWAP
    R> IF NEGATE THEN \ if denominator was negative then negate the remainder
    SWAP
;

: S/REM ( n d -- r q )
    DUP >R ( n d -- ) ( R: -- d )
    SWAP DUP 0< DUP >R ( d n -- ) ( R: -- d nSign )
    IF NEGATE THEN
    SWAP DUP 0< DUP >R ( d n -- ) ( R: -- d nSign dSign )
    IF NEGATE THEN
    U/MOD ( rem quo ) ( R: d nSign dSign )
    SWAP
    R@ IF NEGATE THEN \ negate remainder if denominator was negative

    R> R> <> IF \ if the signs are different
        DUP IF \ if the remainder is not zero
            NEGATE R> + \ add on the demoninator
            SWAP NEGATE 1- \ negate quotient and subtract 1
        ELSE \ remainder is zero
            R> DROP
            SWAP NEGATE \ just negate quotient
        THEN
    ELSE
        R> DROP SWAP
    THEN
;

\ set /MOD to symmetric division by default
DEFER /MOD
' S/REM ' /MOD DEFER!

: / /MOD NIP ;
: MOD /MOD DROP ;

: /_MOD F/MOD ;
: /_ /_MOD NIP ;
: _MOD /_MOD DROP ;

: /-REM S/REM ;
: /- /-REM NIP ;
: -REM /-REM DROP ;

