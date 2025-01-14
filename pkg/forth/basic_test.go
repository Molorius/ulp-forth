/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package forth

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Molorius/ulp-c/pkg/asm"
)

// TestPrimitives has very basic tests for primitive words.
// This is just to rule out initial problems, more thorough tests
// will be done through the standard test suite.
func TestPrimitives(t *testing.T) {
	tests := []struct {
		name   string
		asm    string
		expect string
	}{
		{
			name:   "blank",
			asm:    ": MAIN ESP.DONE ;",
			expect: "",
		},
		{
			name:   "print u16",
			asm:    ": MAIN 123 U. 456 U. ESP.DONE ;",
			expect: "123 456 ",
		},
		{
			name:   "print char",
			asm:    ": MAIN 'A' ESP.PRINTCHAR 'B' ESP.PRINTCHAR 'C' ESP.PRINTCHAR ESP.DONE ;",
			expect: "ABC",
		},
		{
			name:   "double",
			asm:    wrapMain("0xFFFFFF. u. u."),
			expect: "255 65535 ",
		},
		{
			name:   "+",
			asm:    wrapMain("1 2 + u."),
			expect: "3 ",
		},
		{
			name:   "-",
			asm:    wrapMain("3 1 - u."),
			expect: "2 ",
		},
		{
			name:   "LSHIFT",
			asm:    wrapMain("1 0 LSHIFT u. 1 1 LSHIFT u. 1 2 LSHIFT u."),
			expect: "1 2 4 ",
		},
		{
			name:   "SWAP",
			asm:    wrapMain("1 2 SWAP u. u."),
			expect: "1 2 ",
		},
		{
			name:   "DUP",
			asm:    wrapMain("456 789 DUP u. u. u."),
			expect: "789 789 456 ",
		},
		{
			name:   "ROT",
			asm:    wrapMain("1 2 3 ROT u. u. u."),
			expect: "1 3 2 ",
		},
		{
			name:   "DROP",
			asm:    wrapMain("1 2 3 DROP u. u."),
			expect: "2 1 ",
		},
		{ // not really needed but being thorough
			name:   "EXIT",
			asm:    wrapMain("1 U. ESP.DONE EXIT 2 U."),
			expect: "1 ",
		},
		{
			name:   "IF true",
			asm:    wrapMain("TRUE IF 123 THEN U."),
			expect: "123 ",
		},
		{
			name:   "IF false",
			asm:    wrapMain("456 FALSE IF 123 THEN U."),
			expect: "456 ",
		},
		{
			name:   "IF ELSE true",
			asm:    wrapMain("TRUE IF 123 ELSE 456 THEN U."),
			expect: "123 ",
		},
		{
			name:   "IF ELSE false",
			asm:    wrapMain("FALSE IF 123 ELSE 456 THEN U."),
			expect: "456 ",
		},
		{
			name:   ">R R>",
			asm:    wrapMain("123 234 >R U. R> U."),
			expect: "123 234 ",
		},
		{
			name:   "EXECUTE primitive",
			asm:    wrapMain("1 2 ['] + EXECUTE U."),
			expect: "3 ",
		},
		{
			name:   "EXECUTE word",
			asm:    wrapMain("['] FALSE EXECUTE U."),
			expect: "0 ",
		},
		{
			name:   "@ !",
			asm:    "VARIABLE V 789 V ! : MAIN V @ U. 123 456 V ! U. V @ U. ESP.DONE ;",
			expect: "789 123 456 ",
		},
		{
			name:   "PICK",
			asm:    wrapMain("123 456 789 2 PICK u. u. u. u."),
			expect: "123 789 456 123 ",
		},
		{
			name:   "DEPTH",
			asm:    wrapMain("123 DEPTH DEPTH u. u. u."),
			expect: "2 1 123 ",
		},
		{
			name:   "U/MOD",
			asm:    wrapMain("10 2 U/MOD U. U. 123 2 U/MOD U. U."),
			expect: "5 0 61 1 ",
		},
		{
			name:   "NEGATE",
			asm:    wrapMain("-1 NEGATE U. -2 NEGATE U."),
			expect: "1 2 ",
		},
		{
			name:   ".",
			asm:    wrapMain("0 . 1 . 2 . 3 . -1 . -2 . -3 ."),
			expect: "0 1 2 3 -1 -2 -3 ",
		},
		{
			name:   "C@",
			asm:    ": main [ BL WORD test ] LITERAL C@ u. ESP.DONE ;",
			expect: "4 ",
		},
		{
			name:   "BASE",
			asm:    wrapMain("BASE @ U."),
			expect: "10 ",
		},
		{
			name:   "DO",
			asm:    wrapMain("4 0 DO I . LOOP"),
			expect: "0 1 2 3 ",
		},
		{
			name: "tail call",
			asm: `
				\ The point of this isn't to test
				\ a forth algorithm, it's to test that the
				\ compiler is succesfully able to generate
				\ a tail call.
				\ If it doesn't, the return stack pointer
				\ will overflow on the ulp.

				\ countdown from n to 0
				: countdown ( n -- )
					DUP 0= IF
						DROP \ we don't care about the result
					ELSE
						1-
						RECURSE \ tail call recursively!
						EXIT \ TODO get this to work without the EXIT
					THEN
				;

				: MAIN
					DEPTH u. \ print the depth
					5 countdown DEPTH u. \ try after a few recursive calls
					2000 countdown DEPTH u. \ crank it up to 11
					ESP.DONE
				;
			`,
			expect: "0 0 0 ",
		},
	}
	r := asm.Runner{}
	r.SetDefaults()
	err := r.SetupPort()
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runOutputTest(tt.asm, tt.expect, t, &r)
		})
	}
}

func runOutputTest(code string, expected string, t *testing.T, r *asm.Runner) {
	// run the test directly on host
	t.Run("host", func(t *testing.T) {
		parallel(r, t)
		// set up the virtual machine
		var buff bytes.Buffer
		vm := VirtualMachine{Out: &buff}
		err := vm.Setup()
		if err != nil {
			t.Fatalf("failed to set up vm: %s", err)
		}
		// run the code through the interpreter
		err = vm.Execute([]byte(code))
		if err != nil {
			t.Fatalf("failed to execute test code: %s", err)
		}
		// run the test on host
		err = vm.Execute([]byte(" MAIN "))
		if err != nil {
			t.Errorf("error while running: %s", err)
		} else {
			got := buff.String()
			if got != expected {
				t.Errorf("expected \"%s\" got \"%s\"", expected, got)
			}
		}
	})

	// run the test on ulp with token threaded implementation
	t.Run("token threaded", func(t *testing.T) {
		parallel(r, t)
		// set up the virtual machine
		var buff bytes.Buffer
		vm := VirtualMachine{Out: &buff}
		err := vm.Setup()
		if err != nil {
			t.Fatalf("failed to set up vm: %s", err)
		}
		// run the code through the interpreter
		err = vm.Execute([]byte(code))
		if err != nil {
			t.Fatalf("failed to execute test code: %s", err)
		}
		ulp := Ulp{}
		// cross compile "main"
		assembly, err := ulp.BuildAssembly(&vm, "main")
		if err != nil {
			t.Fatalf("failed to generate assembly: %s", err)
		}
		// run the cross compiled test on emulator and hardware
		r.RunTest(t, assembly, expected)
	})

	// run the test on ulp with subroutine threaded implementation
	t.Run("subroutine threaded", func(t *testing.T) {
		parallel(r, t)
		// set up the virtual machine
		var buff bytes.Buffer
		vm := VirtualMachine{Out: &buff}
		err := vm.Setup()
		if err != nil {
			t.Fatalf("failed to set up vm: %s", err)
		}
		// run the code through the interpreter
		err = vm.Execute([]byte(code))
		if err != nil {
			t.Fatalf("failed to execute test code: %s", err)
		}
		ulp := Ulp{}
		// cross compile "main"
		assembly, err := ulp.BuildAssemblySrt(&vm, "main")
		if err != nil {
			t.Fatalf("failed to generate assembly: %s", err)
		}
		// run the cross compiled test on emulator and hardware
		r.RunTest(t, assembly, expected)
	})
}

// Run the tests in parallel.
//
// We cannot run the tests with hardware in
// parallel because each test needs complete access
// to the hardware.
//
// This should be called before any code setup
// so that if parallelism is allowed, we can
// parallelize the entire forth interpret
// and cross compile process.
func parallel(r *asm.Runner, t *testing.T) {
	if !r.PortSet() { // if the hardware port is not set,
		t.Parallel() // then we can safely execute every test in parallel!
	}
}

func wrapMain(code string) string {
	return fmt.Sprintf(" : MAIN %s ESP.DONE ; ", code)
}
