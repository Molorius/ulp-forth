package forth

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/Molorius/ulp-c/pkg/asm"
)

func TestBlank(t *testing.T) {
	forth := ": MAIN ;"
	runOutputTest(forth, "", t)
}

func TestPrintU16(t *testing.T) {
	forth := ": MAIN 123 U. 456 U. ;"
	runOutputTest(forth, "123 456 ", t)
}

func TestPrintChar(t *testing.T) {
	forth := ": MAIN 'A' ESP.PRINTCHAR 'B' ESP.PRINTCHAR 'C' ESP.PRINTCHAR ;"
	runOutputTest(forth, "ABC", t)
}

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
			name:   "double",
			asm:    "0xFFFFFF. u. u.",
			expect: "255 65535 ",
		},
		{
			name:   "+",
			asm:    "1 2 + u.",
			expect: "3 ",
		},
		{
			name:   "-",
			asm:    "3 1 - u.",
			expect: "2 ",
		},
		{
			name:   "SWAP",
			asm:    "1 2 SWAP u. u.",
			expect: "1 2 ",
		},
		{
			name:   "ROT",
			asm:    "1 2 3 ROT u. u. u.",
			expect: "1 3 2 ",
		},
		{
			name:   "DROP",
			asm:    "1 2 3 DROP u. u.",
			expect: "2 1 ",
		},
		{ // not really needed but being thorough
			name:   "EXIT",
			asm:    "1 U. EXIT 2 U.",
			expect: "1 ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runOutputTest(wrapMain(tt.asm), tt.expect, t)
		})
	}
}

func runOutputTest(code string, expected string, t *testing.T) {
	reduce := true // reduce code: we always want to for size!
	var buff bytes.Buffer
	// set up the virtual machine
	vm := VirtualMachine{Out: &buff}
	err := vm.Setup()
	if err != nil {
		t.Fatalf("failed to set up vm: %s", err)
	}
	// run the code through the interpreter
	err = vm.execute([]byte(code))
	if err != nil {
		t.Fatalf("failed to execute test code: %s", err)
	}
	ulp := Ulp{}
	// cross compile "main"
	assembly, err := ulp.BuildAssembly(&vm, "main")
	if err != nil {
		t.Fatalf("failed to generate assembly: %s", err)
	}

	// run the test directly on host
	t.Run("host", func(t *testing.T) {
		err = vm.execute([]byte(" MAIN "))
		got := buff.String()
		if got != expected {
			t.Errorf("expected \"%s\" got \"%s\"", expected, got)
		}
	})
	// run the cross compiled test on emulator and hardware
	asm.RunTest(t, assembly, expected, reduce)
}

func wrapMain(code string) string {
	return fmt.Sprintf(" : MAIN %s ; ", code)
}
