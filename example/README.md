
# Examples

These are small examples showcasing some features of the ULP and
how to access those in ulp-forth. The example_app has some convenience
features such as a serial reader - for a minimal application, see the
program in util/minimal_app.

The example program expects ulp assembly in `example_app/main/out.S`.
This will be compiled by the espressif assembler so the forth should
be compiled with `--assembly`. This is an esp-idf project. Note that
the espressif linker has a bug and so has 12 less bytes available.

`cd` into the example_app directory. To compile the blink example with
subroutine threading, do:
```bash
ulp-forth build --assembly --output main/out.S -r 8164 --subroutine ../blink.f
```
Or if you want to test the local version of the compiler:
```bash
go run ../../main.go build --assembly --output main/out.S -r 8164 --subroutine ../blink.f
```
Then compile and flash it with:
```bash
idf.py build flash monitor
```
