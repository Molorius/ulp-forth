\ This is a bitbanged i2c implementation
\ with clock stretching. It does not support
\ multi-master mode.

\ These need to be set in order to use i2c.
\ Each pin should always have the output set to low in order
\ to pull the bus low. The 'high' state should disable
\ the output and 'low' should enable the output.
\ The 'get' word should read the pin, returning 0 or 1.
DEFER I2C.SDA_HIGH
DEFER I2C.SDA_LOW
DEFER I2C.SDA_GET
DEFER I2C.SCL_HIGH
DEFER I2C.SCL_LOW
DEFER I2C.SCL_GET

\ lower the sda pin if n is 0, else raise it
: __I2C.SDA_SET ( n -- )
    IF
        I2C.SDA_HIGH EXIT
    THEN
    I2C.SDA_LOW
;

\ raise the clock pin
: __I2C.CLOCK_RAISE ( -- )
    I2C.SCL_HIGH \ raise clock
    BEGIN \ loop while clock is stretched (held low)
        I2C.SCL_GET
    UNTIL
;

\ write a single bit to the bus
: __I2C.WRITE_BIT ( n -- )
    __I2C.SDA_SET \ set the bit accordingly
    __I2C.CLOCK_RAISE \ raise clock
    I2C.SCL_LOW \ lower clock
;

\ read a single bit from the bus
: __I2C.READ_BIT ( -- n )
    I2C.SDA_HIGH \ release sda
    __I2C.CLOCK_RAISE \ raise clock
    I2C.SDA_GET \ read the bit
    I2C.SCL_LOW \ lower clock
;

\ send a start condition on the bus
: I2C.START ( -- )
    I2C.SDA_HIGH
    __I2C.CLOCK_RAISE
    I2C.SDA_LOW
    I2C.SCL_LOW
;

\ send a stop condition on the bus
: I2C.STOP ( -- )
    I2C.SCL_LOW
    I2C.SDA_LOW
    __I2C.CLOCK_RAISE
    I2C.SDA_HIGH
;

\ write the byte n to the bus and return true if device sends an ack
: I2C.WRITE ( n -- ack )
    8 0 DO
        DUP 0x80 AND __I2C.WRITE_BIT \ write the highest bit
        1 LSHIFT \ shift the byte up
    LOOP
    DROP \ drop the byte
    __I2C.READ_BIT \ read the ack bit
    0= \ invert the ack, low means acknowledged
;

\ read a byte n from the bus (does not ack/nack)
: I2C.READ ( -- n )
    0 \ create the result
    8 0 DO
        1 LSHIFT \ shift the result
        __I2C.READ_BIT \ read the incoming bit
        OR \ put the bit in the result
    LOOP
;

\ send an ack bit
: I2C.ACK ( -- )
    0 __I2C.WRITE_BIT
;

\ send a nack bit
: I2C.NACK ( -- )
    1 __I2C.WRITE_BIT
;

\ send the start condition and declare a read
: I2C.START_READ ( address -- ack )
    1 LSHIFT \ shift the address
    I2C.START \ send the start condition
    1 OR \ set the read bit
    I2C.WRITE \ write the value, return ack
;

\ send the start condition and declare a write
: I2C.START_WRITE ( address -- ack )
    I2C.START \ send the start condition
    1 LSHIFT \ shift the address
    I2C.WRITE \ write the value, return ack
;
