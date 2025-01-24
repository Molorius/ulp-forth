/*
Copyright 2025 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/Molorius/ulp-c/pkg/usb"
	"github.com/spf13/cobra"
)

var flashTestCmd = &cobra.Command{
	Use:   "flash_test",
	Short: "Flash the test application",
	Long: `flash_test is used to write the test application
to the flash of an esp32. The test application can be used
for unit tests or for general debugging of the ULP.

It requires the serial port that the esp32 is connected to.
Requires that esptool.py is installed.

Example:
ulp-forth flash_test /dev/tty.usbserial-0001`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			os.Exit(0)
		}
		if len(args) != 1 {
			fmt.Printf("only port expected, %d arguments passed in\r\n", len(args))
			os.Exit(1)
		}

		port := args[0]
		err := usb.WriteTestApp(port)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(flashTestCmd)
}
