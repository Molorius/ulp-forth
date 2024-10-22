package cmd

import (
	"fmt"
	"os"

	"github.com/Molorius/ulp-forth/pkg/forth"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute the input forth files and set up an interpreter",
	Long: `Executes the input forth files then sets up in interpreter
for testing. This runs purely on the host device.

Examples:
ulp-forth run
ulp-forth run file1.f`,
	Run: func(cmd *cobra.Command, args []string) {
		vm := forth.VirtualMachine{}
		err := vm.Setup()
		if err != nil {
			fmt.Println(err)
			return
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
				return
			}
			vm.ExecuteFile(f)
			f.Close()
		}
		fmt.Fprintln(vm.Out, "ulp-forth")
		for {
			err = vm.Repl()
			if err != nil {
				if err.Error() == "Interrupt" {
					fmt.Println()
					return
				}
				fmt.Println(err)
				vm.Reset()
			}
			state, err := vm.State.Get()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			if state == uint16(forth.StateExit) {
				return
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
