
\ Initializes the led. Change to whichever gpio you
\ want to use.
\ You can save space by initializing this with the esp32.
: led_init ( -- )
    gpio2.enable
    gpio2.output_enable
    \ or use the rtc_gpio equivalents:
    \ rtc_gpio12.enable
    \ rtc_gpio12.output_enable
;

\ Sets the led gpio to high or low.
\ Takes in a value. Set low if 0 else set high.
\ Change to whichever gpio you want to use.
: led_set ( val -- )
    gpio2.set
    \ or use the rtc_gpio equivalent:
    \ rtc_gpio12.set
;

\ The number of milliseconds we want to delay.
1000 constant ms

: main
    led_init
    begin
        \ set high
        1 led_set
        ms delay_ms
        \ set low
        0 led_set
        ms delay_ms
    again
;
