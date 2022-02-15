package cli

import (
	"fmt"
	"os"

	"github.com/alranel/arduino-testlib/internal/configuration"
	"github.com/alranel/arduino-testlib/internal/report"
	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:   "report --datadir /path/to/dir",
	Short: "Generate a report",
	Long:  `This command reads all the stored test results and generates a report`,
	Run:   runReport,
}

func init() {
	reportCmd.PersistentFlags().StringP("output", "o", "report", "The directory to write the HTML report to.")
	rootCmd.AddCommand(reportCmd)
}

func runReport(cmd *cobra.Command, cliArguments []string) {
	// Read configuration
	if err := configuration.Initialize(cmd.Flags()); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Check if a --datadir was supplied
	datadirPath, _ := cmd.Flags().GetString("datadir")
	if datadirPath == "" {
		fmt.Fprintf(os.Stderr, "Missing required --datadir option\n")
		os.Exit(1)
	}

	outputDir, _ := cmd.Flags().GetString("output")

	report.Generate(datadirPath, outputDir)
}
