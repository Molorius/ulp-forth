\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

: DNEGATE 0. 2SWAP D- ;
: DABS DUP 0< IF DNEGATE THEN ;
: M+ S>D D+ ;

: D= ( x1 x2 y1 y2 -- bool )
    ROT ( x1 y1 y2 x2 )
    <> IF \ if not equal then drop the rest
        2DROP FALSE EXIT
    THEN
    = \ return equality
;

: DU<
    ROT SWAP \ move the high cells to the top
    2DUP <> IF \ if the high cells are not equal
        2SWAP \ then ignore the low cells
    THEN \ drop the top
    2DROP
    U< \ compare the bottom
;

: DU>
    2SWAP DU<
;

: D0=
    OR 0=
;

: D0<
    NIP 0<
;

: D0<>
    OR 0<>
;

: D0>
    NIP 0>
;

: D<>
    D= 0=
;

: D>
    0x80000000. D-
    2SWAP 0x80000000. D-
    DU<
;

: D<
    2SWAP D>
;

: D>S
    DROP
;

: DMAX 2OVER 2OVER D< IF 2SWAP THEN 2DROP ;
: DMIN 2OVER 2OVER D> IF 2SWAP THEN 2DROP ;

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
: GLOBAL-2VARIABLE 2 GLOBAL-BUFFER: ;

: 2VALUE ( x y "<spaces>name" -- )
    2 ALLOCATE DROP ( x y addr )
    SWAP OVER 1+ ( x addr y addr+1 )
    ! ( x addr )
    SWAP OVER ( addr x addr )
    ! ( addr )
    DUP 1+ SWAP ( addr+1 addr )
    :
    POSTPONE LITERAL
    POSTPONE @
    POSTPONE LITERAL
    POSTPONE @
    POSTPONE ;
;

: 2ROT
    5 ROLL 5 ROLL
;
