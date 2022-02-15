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
	rootCmd.AddCommand(testCmd)
}

func runTest(cmd *cobra.Command, cliArguments []string) {
	if err := configuration.Initialize(cmd.Flags()); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	cliclient.Init()
	cliclient.InstallCores()

	var tr test.TestResults
	tr = test.TestLib(cliArguments[0], tr)

	b, _ := json.MarshalIndent(tr, "", "  ")
	fmt.Printf("%s\n", b)
}