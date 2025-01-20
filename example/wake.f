\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

1000 constant ms

: main
    begin
        wake \ send the wake signal to the esp32
        ms delay_ms \ delay
    again
;
