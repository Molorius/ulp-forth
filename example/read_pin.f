\ Put an led on the output pin,
\ and a button that will connect to ground
\ on the input pin.

\ Initializes the pins.
\ This can be done on the esp32 to save ulp save.
: pin_init
    \ output
    gpio2.enable
    gpio2.output_enable
    \ input
    gpio32.enable
    gpio32.input_enable
    gpio32.pullup_enable \ use internal pullup resistors
;

\ Invert the value to the logical inverse !a
\ so we can convert the pulled up input to
\ our led output.
: ! ( a -- !a )
    0= \ if 0 then this outputs -1, otherwise it outputs 0
;

: main
    pin_init

    begin
        gpio32.get \ get the value of the pin
        ! \ invert the value
        gpio2.set \ light the led if the button is pressed
    again
;
