package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alranel/arduino-testlib/internal/cliclient"
	"github.com/alranel/arduino-testlib/internal/configuration"
	"github.com/alranel/arduino-testlib/pkg/test"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
	"golang.org/x/sync/semaphore"
)

var testallCmd = &cobra.Command{
	Use:   "testall --datadir /path/to/dir",
	Short: "Test all libraries",
	Long:  `This command performs an incremental test of all libraries`,
	Run:   runTestall,
}

func init() {
	testallCmd.PersistentFlags().IntP("threads", "j", 1, "How many parallel jobs to run")
	testallCmd.PersistentFlags().BoolP("force", "f", false, "Re-test all library-core combinations even if already seen")
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

	instance := cliclient.NewInstance()

	// Install all the required cores
	instance.InstallCores()

	// Define the list of the libraries to test. If no libraries were supplied as
	// arguments, the entire list from the Library Registry will be used.
	libraries := make(map[string]string) // unsanitized name => version
	{
		var libs []string
		if len(cliArguments) == 0 {
			libs = instance.GetInstalledLibraries()
		} else {
			// Parse arguments as glob patterns, allowing filters such as "Arduino_*"
			for _, arg := range cliArguments {
				g := glob.MustCompile(arg)
				for _, lib := range instance.GetInstalledLibraries() {
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
	ctx := context.TODO()
	sem := semaphore.NewWeighted(1)
	var done int32
	t0 := time.Now()

	force, _ := cmd.Flags().GetBool("force")
	worker := func(wg *sync.WaitGroup, workerId int) {
		// Create a new CLI instance for each worker, preventing concurrency
		if err := sem.Acquire(ctx, 1); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to acquire semaphore: %v", err)
			os.Exit(1)
		}
		fmt.Printf("[#%d] Initializing CLI\n", workerId)
		instance := cliclient.NewInstance()
		fmt.Printf("[#%d] Done initializing CLI\n", workerId)
		sem.Release(1)

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

			tr = test.TestLibByName(lib, tr, force, instance)

			// Write test results to datadir
			{
				jsonData, _ := json.MarshalIndent(tr, "", "  ")
				err := ioutil.WriteFile(testResultsFile, jsonData, 0644)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not save test results: %v\n", err)
					os.Exit(1)
				}
			}

			// Increment counter and print stats
			atomic.AddInt32(&done, 1)
			eta := int(time.Now().Sub(t0).Seconds() / float64(done) * float64(len(libraries)-int(done)))
			fmt.Printf("[#%d] done %d/%d libs (ETA: %ds)\n", workerId, done, len(libraries), eta)
		}
	}

	noOfWorkers, _ := cmd.Flags().GetInt("threads")
	if noOfWorkers > 1 {
		fmt.Printf("Warning: the --threads option is experimental\n")
	}
	var wg sync.WaitGroup
	for i := 0; i < noOfWorkers; i++ {
		wg.Add(1)
		go worker(&wg, i)
	}

	// Sort libraries alphabetically
	libNames := make([]string, 0, len(libraries))
	for lib := range libraries {
		libNames = append(libNames, lib)
	}
	sort.Strings(libNames)
	fmt.Printf("Total libraries: %d\n", len(libNames))
	for _, lib := range libNames {
		jobs <- lib
	}
	close(jobs)

	wg.Wait()
}
