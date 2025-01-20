\ Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com
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

\ create a word called tx for serial
13 gpio_number_to_rtc serial.write_9600_baud serial.write_create tx

\ use our tx word for printing
' tx is emit

: main
    pin_init

    ." Starting, this will only print once" cr
    0
    begin
        ." Halting "
        dup u. 1+ \ print out the number and increment it
        cr \ print a carriage return
        halt \ halt execution
        \ execution continues here after the ulp starts again,
        \ so it will immediately loop after starting up
    again
;

