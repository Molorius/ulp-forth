\ Copyright 2024 Blake Felt blake.w.felt@gmail.com
\ This Source Code Form is subject to the terms of the Mozilla Public
\ License, v. 2.0. If a copy of the MPL was not distributed with this
\ file, You can obtain one at https://mozilla.org/MPL/2.0/.

\ Initializes the serial out pin.
\ This can be done on the esp32 to save ulp space.
: pin_init
    gpio13.enable
    gpio13.output_enable
    gpio13.set_high
;

1000 constant ms

: main
    
    begin
        wake \ send the wake signal to the esp32
        ms delay_ms
    again
;
