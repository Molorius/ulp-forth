\ Copyright 2024 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
: CR 10 13 EMIT EMIT ;

: U.NOSPACE ( u -- )
    BASE @ U/MOD \ divide with remainder
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

: TYPE ( c-addr u -- )
    0 ?DO
        DUP C@ \ get the current character
        EMIT   \ emit it
        CHAR+  \ go to next character
    LOOP
    DROP \ remove c-addr
;

: ." POSTPONE S" POSTPONE TYPE ; IMMEDIATE

\ set EMIT to the system printchar by default
' ESP.PRINTCHAR IS EMIT
