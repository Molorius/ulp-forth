package forth

import (
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
			name:   "+",
			asm:    "1 2 + u.",
			expect: "3 ",
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
		f := " : MAIN " + tt.asm + " ; " // put the code inside of a word
		t.Run(tt.name, func(t *testing.T) {
			runOutputTest(f, tt.expect, t)
		})
	}
}

func runOutputTest(code string, expected string, t *testing.T) {
	reduce := true // reduce code: we always want to for size!
	vm := VirtualMachine{}
	err := vm.Setup()
	if err != nil {
		t.Fatalf("failed to set up vm: %s", err)
	}
	err = vm.execute([]byte(code))
	if err != nil {
		t.Fatalf("failed to execute test code: %s", err)
	}
	ulp := Ulp{}
	assembly, err := ulp.BuildAssembly(&vm, "main")
	if err != nil {
		t.Fatalf("failed to generate assembly: %s", err)
	}
	// run the test!
	asm.RunTest(t, assembly, expected, reduce)
}
