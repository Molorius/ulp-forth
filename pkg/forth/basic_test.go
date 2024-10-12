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
			name:   "EXECUTE",
			asm:    wrapMain("['] U. 123 SWAP EXECUTE"),
			expect: "123 ",
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
			name:   "/MOD",
			asm:    wrapMain("10 2 /MOD U. U. 123 2 /MOD U. U."),
			expect: "5 0 61 1 ",
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
	r.RunTest(t, assembly, expected)
}

func wrapMain(code string) string {
	return fmt.Sprintf(" : MAIN %s ESP.DONE ; ", code)
}
