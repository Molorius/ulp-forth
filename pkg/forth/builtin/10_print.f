
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
: U. ( n -- ) ESP.PRINTU16 ;
