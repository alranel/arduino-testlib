package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"sync"

	"github.com/alranel/arduino-testlib/internal/cliclient"
	"github.com/alranel/arduino-testlib/internal/configuration"
	"github.com/alranel/arduino-testlib/pkg/test"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

var testallCmd = &cobra.Command{
	Use:   "testall --datadir /path/to/dir",
	Short: "Test all libraries",
	Long:  `This command performs an incremental test of all libraries`,
	Run:   runTestall,
}

func init() {
	testallCmd.PersistentFlags().IntP("threads", "j", 1, "How many parallel jobs to run")
	rootCmd.AddCommand(testallCmd)
}

func runTestall(cmd *cobra.Command, cliArguments []string) {
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

	cliclient.Init()

	// Install all the required cores
	cliclient.InstallCores()

	// Define the list of the libraries to test. If no libraries were supplied as
	// arguments, the entire list from the Library Registry will be used.
	libraries := make(map[string]string) // unsanitized name => version
	{
		var libs []string
		if len(cliArguments) == 0 {
			libs = cliclient.GetInstalledLibraries()
		} else {
			// Parse arguments as glob patterns, allowing filters such as "Arduino_*"
			for _, arg := range cliArguments {
				g := glob.MustCompile(arg)
				for _, lib := range cliclient.GetInstalledLibraries() {
					if g.Match(lib) {
						libs = append(libs, lib)
					}
				}
			}
		}
		for _, lib := range libs {
			t := strings.SplitN(lib, "@", 2)
			version := ""
			if len(t) > 1 {
				version = t[1]
			}
			libraries[t[0]] = version
		}
	}

	var jobs = make(chan string)

	worker := func(wg *sync.WaitGroup) {
		for {
			lib, more := <-jobs
			if !more {
				wg.Done()
				return
			}

			var tr test.TestResults

			// Read previous test results from datadir
			testResultsFile := path.Join(datadirPath, utils.SanitizeName(lib)+".json")
			test.ReadResultsFile(testResultsFile, &tr)

			tr = test.TestLib(lib, tr)

			// Write test results to datadir
			{
				jsonData, _ := json.MarshalIndent(tr, "", "  ")
				err := ioutil.WriteFile(testResultsFile, jsonData, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not save test results: %v\n", err)
					os.Exit(1)
				}
			}
		}
	}

	noOfWorkers, _ := cmd.Flags().GetInt("threads")
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go worker(&wg)
	}

	for lib := range libraries {
		jobs <- lib
	}
	close(jobs)

	wg.Wait()
}
