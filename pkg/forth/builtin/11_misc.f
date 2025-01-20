\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

: HIDE ( xt -- )
    TRUE SWAP SET-HIDDEN
;

: UNHIDE ( xt -- )
    FALSE SWAP SET-HIDDEN
;

\ constants used for defining token threaded primitives
0 CONSTANT TOKEN_NEXT_NONSTANDARD
1 CONSTANT TOKEN_NEXT_NORMAL
2 CONSTANT TOKEN_NEXT_SKIP_R2
3 CONSTANT TOKEN_NEXT_SKIP_LOAD
