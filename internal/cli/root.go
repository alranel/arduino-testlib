package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:                   "arduino-testlib",
	Short:                 "Compile an Arduino library on multiple platforms.",
	Long:                  "This tool tries to compile an Arduino library on multiple platforms in order to check its compatibility.",
	DisableFlagsInUseLine: true,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().String("datadir", "", "The directory where test results are stored.")
	rootCmd.PersistentFlags().String("cli-datadir", "", "A custom directory for arduino-cli data.")
	rootCmd.PersistentFlags().String("additional-urls", "", "Comma-separated list of additional URLs for the Boards Manager.")
	rootCmd.PersistentFlags().StringSlice("fqbn", []string{}, "The FQBN(s) to compile the library against.")
}

// Execute starts the cobra command parsing chain.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
