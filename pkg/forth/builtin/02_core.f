
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

\ used for testing, will remove later
: main 123 u. ;
bye
