
: D= ( x1 x2 y1 y2 -- bool )
    ROT ( x1 y1 y2 x2 )
    <> IF \ if not equal then drop the rest
        2DROP FALSE EXIT
    THEN
    = \ return equality
;

: DU<
    ROT
    2DUP = IF
        2DROP U<
    ELSE
        U> >R 2DROP R>
    THEN
;

: DU>
    ROT
    2DUP = IF
        2DROP U>
    ELSE
        U< >R 2DROP R>
    THEN
;

: D0=
    OR 0=
;

: D0<
    NIP 0<
;

: D>S
    DROP
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
