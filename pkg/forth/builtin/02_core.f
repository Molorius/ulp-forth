
\ ['] parses the next name and compiles the execution token of that name. Immediate.
: ['] ( -- ) ' POSTPONE LITERAL ; IMMEDIATE

\ RECURSE compiles the most recently defined name, which usually means the current definition. 
\ Immediate.
: RECURSE ( -- ) LAST COMPILE, ; IMMEDIATE

: ESP.FUNC.TYPE.ACK 0 ;
: ESP.FUNC.TYPE.DONE 1 ;
: ESP.FUNC.TYPE.PRINTU16 2 ;
: ESP.FUNC.TYPE.PRINTCHAR 3 ;
: ESP.FUNC ( value function_number -- )
    MUTEX.TAKE ESP.FUNC.UNSAFE MUTEX.GIVE
    \ we don't have branches yet, pause instead of checking for ack
    DEBUG.PAUSE DEBUG.PAUSE
;
: ESP.PRINTU16 ( n -- ) ESP.FUNC.TYPE.PRINTU16 ESP.FUNC ;
: ESP.PRINTCHAR ( char -- ) ESP.FUNC.TYPE.PRINTCHAR ESP.FUNC ;
: ESP.DONE ( -- ) 0 ESP.FUNC.TYPE.DONE ESP.FUNC ;
: U. ( n -- ) ESP.PRINTU16 ;

: NIP ( a b -- b ) SWAP DROP ;
: 1+ ( x -- x+1 ) 1 + ;
: 1- ( x -- x-1 ) 1 - ;
\ Negate x. -2 becomes 2, 3 becomes -3, etc.
: NEGATE ( x -- -x ) 0 SWAP - ;
\ Invert all bits. 0xFFFF becomes 0, 0xFFF0 becomes 0x000F, etc.
: INVERT ( x -- x^0xFFFF ) NEGATE 1- ;

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
