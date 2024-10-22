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
		for _, arg := range args {
			f, err := os.Open(arg)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = vm.ExecuteFile(f)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			f.Close()
		}
		ulp := forth.Ulp{}
		assembly, err := ulp.BuildAssembly(&vm, "MAIN")
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
		f.Write(out)
	},
}

func init() {
	rootCmd.AddCommand(buildCmd)

	buildCmd.Flags().String(CmdOutput, "", "Name of output file")
	buildCmd.Flags().IntP(CmdReserved, "r", 8176, "Number of reserved bytes for the ULP, for use with --assembly flag")

	buildCmd.Flags().Bool(CmdAssembly, false, "Output assembly that can be compiled by the main assemblers, set the --reserved flag before using")
	buildCmd.Flags().Bool(CmdCustomAssembly, false, "Output assembly only for use by ulp-asm")
	buildCmd.MarkFlagsMutuallyExclusive(CmdCustomAssembly, CmdAssembly)
}
