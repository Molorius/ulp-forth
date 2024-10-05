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
