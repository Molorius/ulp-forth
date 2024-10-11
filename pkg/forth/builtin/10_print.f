
: ESP.FUNC.TYPE.ACK 0 ;
: ESP.FUNC.TYPE.DONE 1 ;
: ESP.FUNC.TYPE.PRINTU16 2 ;
: ESP.FUNC.TYPE.PRINTCHAR 3 ;

: ESP.FUNC.READ ( -- function_number )
    MUTEX.TAKE ESP.FUNC.READ.UNSAFE MUTEX.GIVE ;

: ESP.FUNC ( value function_number -- )
    \ we don't do anything with the read other than
    \ checking for 0 and we immediately take the mutex,
    \ so it's safe (probably)
    BEGIN ESP.FUNC.READ.UNSAFE 0= UNTIL \ loop until previous is not 0
    MUTEX.TAKE ESP.FUNC.UNSAFE MUTEX.GIVE \ write!
;

: ESP.PRINTU16 ( n -- ) ESP.FUNC.TYPE.PRINTU16 ESP.FUNC ;
: ESP.PRINTCHAR ( char -- ) ESP.FUNC.TYPE.PRINTCHAR ESP.FUNC ;
: ESP.DONE ( -- ) 0 ESP.FUNC.TYPE.DONE ESP.FUNC ;

DEFER EMIT \ let us change EMIT
: SPACE BL EMIT ;

: U.NOSPACE ( u -- )
    BASE @ /MOD \ divide with remainder
    ?DUP IF RECURSE THEN \ if the quotient is nonzero, print that first
    DUP #10 U< IF \ if it's 0-9 then we want those characters
        '0'
    ELSE
        [ 'A' 10 - ] LITERAL \ otherwise start printing at 'A' character
    THEN
    + EMIT \ add on the character offset and print it!
;

: U. ( u -- ) U.NOSPACE SPACE ;
: .  ( n -- )
    DUP 0< IF \ if the value is less than 0
        '-' EMIT \ print a minus sign
        NEGATE \ and negate the value
    THEN
    U. \ print the value
;

\ set EMIT to the system printchar by default
' ESP.PRINTCHAR ' EMIT DEFER!
