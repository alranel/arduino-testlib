package cli

import (
	"fmt"
	"os"

	"github.com/alranel/arduino-testlib/internal/cliclient"
	"github.com/alranel/arduino-testlib/internal/configuration"
	"github.com/spf13/cobra"
)

var installallCmd = &cobra.Command{
	Use:   "installall",
	Short: "Install all the Arduino libraries",
	Long:  `This command performs a batch install/upgrade of all the libraries in the Arduino Library Manager`,
	Run:   runInstallall,
}

func init() {
	rootCmd.AddCommand(installallCmd)
}

func runInstallall(cmd *cobra.Command, cliArguments []string) {
	// Read configuration
	if err := configuration.Initialize(cmd.Flags()); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	cliclient.Init()

	for _, lib := range cliclient.GetAllLibraries() {
		// Use the unsanitized name to install the library
		cliclient.InstallLibrary(lib, "")
	}
}
