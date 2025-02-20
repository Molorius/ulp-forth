/*
Copyright 2024-2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Molorius/ulp-c/pkg/asm"
	"github.com/Molorius/ulp-forth/pkg/forth"
	"github.com/spf13/cobra"
)

const CmdOutput = "output"
const CmdReserved = "reserved"
const CmdAssembly = "assembly"
const CmdCustomAssembly = "custom_assembly"
const CmdSubroutineThreading = "subroutine"

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the forth code",
	Long: `Executes the input forth files then cross compiles
the "MAIN" word as executable ULP code.

Example:
ulp-forth build --assembly --reserved 1024 file1.f file2.f`,
	Run: func(cmd *cobra.Command, args []string) {
		vm := forth.VirtualMachine{}
		err := vm.Setup()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = vm.BuiltinEsp32()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			defer f.Close()
			err = vm.ExecuteFile(f)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		ulp := forth.Ulp{}
		var assembly string
		subroutine, _ := cmd.Flags().GetBool(CmdSubroutineThreading)
		if subroutine {
			assembly, err = ulp.BuildAssemblySrt(&vm, "MAIN")
		} else {
			assembly, err = ulp.BuildAssembly(&vm, "MAIN")
		}
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		buildAssembly, _ := cmd.Flags().GetBool(CmdAssembly)
		buildCustomAsm, _ := cmd.Flags().GetBool(CmdCustomAssembly)
		output, _ := cmd.Flags().GetString(CmdOutput)
		reduce := true // always reduce
		reserved, _ := cmd.Flags().GetInt(CmdReserved)
		assembler := asm.Assembler{}
		var out []byte
		if buildAssembly {
			if output == "" {
				output = "out.S"
			}

			built, err := assembler.BuildAssembly(assembly, "forth.S", reserved, reduce)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			out = built
		} else if buildCustomAsm {
			if output == "" {
				output = "out.nonportable.S"
			}
			out = []byte(assembly)
		} else { // build the binary
			if output == "" {
				output = "out.bin"
			}
			built, err := assembler.BuildFile(assembly, "forth.S", reserved, reduce)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			out = built
		}

		f, err := os.Create(output)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		defer f.Close()
		f.Write(out)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().String(CmdOutput, "", "Name of the output file.")
	buildCmd.Flags().IntP(CmdReserved, "r", 8176, "Number of reserved bytes for the ULP, for use with --assembly flag. Note that the espressif linker reserves an extra 12 bytes.")

	buildCmd.Flags().Bool(CmdAssembly, false, "Output assembly that can be compiled by the main assemblers, set the --reserved flag before using.")
	buildCmd.Flags().Bool(CmdCustomAssembly, false, "Output assembly only for use by ulp-asm, another project by this author.")
	buildCmd.MarkFlagsMutuallyExclusive(CmdCustomAssembly, CmdAssembly)

	buildCmd.Flags().Bool(CmdSubroutineThreading, false, "Use the subroutine threading model. Faster but larger.")
}
