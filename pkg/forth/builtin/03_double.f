
: D= ( x1 x2 y1 y2 -- bool )
    ROT ( x1 y1 y2 x2 )
    <> IF \ if not equal then drop the rest
        2DROP FALSE EXIT
    THEN
    = \ return equality
;

: 2LITERAL
    SWAP
    POSTPONE LITERAL
    POSTPONE LITERAL
; IMMEDIATE

: 2CONSTANT
    :
    POSTPONE 2LITERAL
    POSTPONE ;
;

: 2VARIABLE 2 BUFFER: ;
