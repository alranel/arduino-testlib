package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/alranel/arduino-testlib/internal/cliclient"
	"github.com/alranel/arduino-testlib/internal/configuration"
	"github.com/alranel/arduino-testlib/pkg/test"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test LIB",
	Short: "Test a single library",
	Long:  `This command tests a single library on the specified boards`,
	Run:   runTest,
}

func init() {
	testCmd.PersistentFlags().BoolP("force", "f", false, "Re-test all library-core combinations even if already seen")
	rootCmd.AddCommand(testCmd)
}

func runTest(cmd *cobra.Command, cliArguments []string) {
	if err := configuration.Initialize(cmd.Flags()); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}
	if len(cliArguments) != 1 {
		fmt.Fprintf(os.Stderr, "Invalid arguments: please supply the path to a library to test\n")
		os.Exit(1)
	}

	instance := cliclient.NewInstance()
	instance.InstallCores()

	var tr test.TestResults
	force, _ := cmd.Flags().GetBool("force")
	tr = test.TestLib(cliArguments[0], tr, force, instance)

	b, _ := json.MarshalIndent(tr, "", "  ")
	fmt.Printf("%s\n", b)
}
