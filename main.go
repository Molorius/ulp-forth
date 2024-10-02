package main

import (
	"fmt"

	"github.com/Molorius/ulp-forth/pkg/forth"
)

func main() {
	vm := forth.VirtualMachine{}
	err := vm.Setup()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = vm.Repl()
	if err != nil {
		fmt.Println(err)
	}
}
