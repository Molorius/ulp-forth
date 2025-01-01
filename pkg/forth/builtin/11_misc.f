\ Copyright 2024 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

: HIDE ( xt -- )
    TRUE SWAP SET-HIDDEN
;

: UNHIDE ( xt -- )
    FALSE SWAP SET-HIDDEN
;
