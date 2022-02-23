package report

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/alranel/arduino-testlib/internal/util"
	"github.com/alranel/arduino-testlib/pkg/test"
	"github.com/arduino/arduino-cli/arduino/utils"
	"golang.org/x/mod/semver"
)

func Generate(datadirPath string, outputDir string) {
	// Prepare the data structures
	libraries := make(map[string]string)
	boards := make(map[string]map[string]bool)
	type libBoardPair struct{ lib, board string }
	type compatibilityStatus string
	const (
		PASS_CLAIM   compatibilityStatus = "PASS_CLAIM"
		PASS_NOCLAIM compatibilityStatus = "PASS_NOCLAIM"
		FAIL_CLAIM   compatibilityStatus = "FAIL_CLAIM"
		FAIL_NOCLAIM compatibilityStatus = "FAIL_NOCLAIM"
	)
	compatibility := make(map[libBoardPair]compatibilityStatus)
	compatibilityAsterisk := make(map[string]bool) // lib => has_asterisk
	testResults := make(map[libBoardPair]test.TestResult)
	numExamples := make(map[int]int)

	// Read library data
	files, _ := ioutil.ReadDir(datadirPath)
	for _, file := range files {
		// Read test results file
		var tr test.TestResults
		if !test.ReadResultsFile(path.Join(datadirPath, file.Name()), &tr) {
			continue
		}

		// Sort tests by lib version and core version
		// so that we override older data with newer data
		sort.Slice(tr.Tests, func(i, j int) bool {
			if semver.Compare(tr.Tests[i].Version, tr.Tests[j].Version) == -1 {
				return true
			} else if semver.Compare(tr.Tests[i].CoreVersion, tr.Tests[j].CoreVersion) == -1 {
				return true
			}
			return false
		})

		// Go through all the tests
		nEx := 0
		for _, t := range tr.Tests {
			libraries[tr.Name] = t.Version

			if t.CoreVersion == "" {
				fmt.Printf("EMPTY CORE VERSION! %s\n", tr.Name)
			}

			if boards[t.FQBN] == nil {
				boards[t.FQBN] = make(map[string]bool)
			}
			boards[t.FQBN][t.CoreVersion] = true
			nEx = len(t.Examples)

			// Find the compatibility status
			var cSt compatibilityStatus
			if t.Result == test.PASS {
				if util.CoreInArchitectures(t.Core, t.Architectures) {
					cSt = PASS_CLAIM
				} else {
					cSt = PASS_NOCLAIM
				}
			} else {
				if util.CoreInArchitectures(t.Core, t.Architectures) {
					cSt = FAIL_CLAIM
				} else {
					cSt = FAIL_NOCLAIM
				}
			}
			compatibility[libBoardPair{tr.Name, t.FQBN}] = cSt
			testResults[libBoardPair{tr.Name, t.FQBN}] = t
		}
		if len(tr.Tests) > 0 {
			for _, arch := range tr.Tests[len(tr.Tests)-1].Architectures {
				if arch == "*" {
					compatibilityAsterisk[tr.Name] = true
				}
			}
		}
		numExamples[nEx] = numExamples[nEx] + 1
	}

	// Compute statistics
	numLibs := len(libraries)
	percent := func(n int) string { return fmt.Sprintf("%.1f%%", float32(n)/float32(numLibs)*100) }

	type boardReportData struct {
		Name, Architecture, Versions                                            string
		Claim, ExplicitClaim, ClaimMismatch, Pass, Fail, Untested               int
		PassClaim, PassNoClaim, FailClaim, FailClaimAsterisk, FailExplicitClaim int
	}
	type libraryReportData struct {
		Name, ReportFile, Version, URL string
		BoardCompatibility             map[string]compatibilityStatus
		BoardTestResults               map[string]test.TestResult
		Examples                       []string
	}
	type exampleReportData struct {
		Num, Count int
	}
	reportData := struct {
		Timestamp                                   string
		NumLibs, NumBoards                          int
		NumLibsCompatibilityAsterisk                int
		NumLibsClaimNoBoards, NumLibsClaimAllBoards int
		NumLibsPassNoBoards, NumLibsPassAllBoards   int
		NumLibsFailClaim                            int
		NumLibsAsteriskFail                         int
		HasUntested                                 bool
		Boards                                      []boardReportData
		Examples                                    []exampleReportData
		Libraries                                   []libraryReportData
	}{
		Timestamp:                    time.Now().Format(time.RFC850),
		NumLibs:                      numLibs,
		NumBoards:                    len(boards),
		NumLibsCompatibilityAsterisk: len(compatibilityAsterisk),
	}

	// Board statistics
	for board := range boards {
		var versions []string
		for v := range boards[board] {
			versions = append(versions, v)
		}
		semver.Sort(versions)

		c := boardReportData{
			Name:         board,
			Architecture: util.ArchitectureFromFQBN(board),
			Versions:     strings.Join(versions, ","),
		}

		cnt := make(map[compatibilityStatus]int)
		failClaimAsterisk := 0
		for pair, result := range compatibility {
			if pair.board == board {
				cnt[result] = cnt[result] + 1
				if result == FAIL_CLAIM && compatibilityAsterisk[pair.lib] {
					failClaimAsterisk = failClaimAsterisk + 1
				}
			}
		}
		c.Claim = cnt[PASS_CLAIM] + cnt[FAIL_CLAIM]
		c.ExplicitClaim = c.Claim - reportData.NumLibsCompatibilityAsterisk
		c.ClaimMismatch = cnt[PASS_NOCLAIM] + cnt[FAIL_CLAIM]
		c.Pass = cnt[PASS_CLAIM] + cnt[PASS_NOCLAIM]
		c.Fail = cnt[FAIL_CLAIM] + cnt[FAIL_NOCLAIM]
		c.Untested = numLibs - (c.Pass + c.Fail)
		c.PassClaim = cnt[PASS_CLAIM]
		c.PassNoClaim = cnt[PASS_NOCLAIM]
		c.FailClaim = cnt[FAIL_CLAIM]
		c.FailClaimAsterisk = failClaimAsterisk
		c.FailExplicitClaim = c.FailClaim - c.FailClaimAsterisk
		if c.Untested > 0 {
			reportData.HasUntested = true
		}

		reportData.Boards = append(reportData.Boards, c)
	}
	sort.Slice(reportData.Boards, func(i, j int) bool {
		return reportData.Boards[i].Name < reportData.Boards[j].Name
	})

	// Library statistics
	var libNames []string
	for lib := range libraries {
		libNames = append(libNames, lib)
	}
	sort.Strings(libNames)
	for _, lib := range libNames {
		lData := libraryReportData{
			Name:               lib,
			ReportFile:         utils.SanitizeName(lib) + ".html",
			URL:                libraryURL(lib),
			Version:            libraries[lib],
			BoardCompatibility: make(map[string]compatibilityStatus),
			BoardTestResults:   make(map[string]test.TestResult),
		}
		totClaim := 0
		totFailClaim := 0
		totPass := 0
		totFail := 0
		for pair, result := range compatibility {
			if pair.lib == lib {
				lData.BoardCompatibility[pair.board] = result
				if result == PASS_CLAIM || result == FAIL_CLAIM {
					totClaim = totClaim + 1
				}
				if result == FAIL_CLAIM {
					totFailClaim = totFailClaim + 1
				}
				if result == PASS_CLAIM || result == PASS_NOCLAIM {
					totPass = totPass + 1
				}
				if result == FAIL_CLAIM || result == FAIL_NOCLAIM {
					totFail = totFail + 1
				}
			}
		}
		exampleNames := make(map[string]bool)
		for pair, t := range testResults {
			if pair.lib == lib {
				lData.BoardTestResults[pair.board] = t
				for _, e := range t.Examples {
					exampleNames[e.Name] = true
				}
			}
		}
		for e := range exampleNames {
			lData.Examples = append(lData.Examples, e)
		}
		reportData.Libraries = append(reportData.Libraries, lData)
		if totClaim == len(boards) {
			reportData.NumLibsClaimAllBoards = reportData.NumLibsClaimAllBoards + 1
		}
		if totClaim == 0 {
			reportData.NumLibsClaimNoBoards = reportData.NumLibsClaimNoBoards + 1
		}
		if totPass == len(boards) {
			reportData.NumLibsPassAllBoards = reportData.NumLibsPassAllBoards + 1
		}
		if totPass == 0 {
			reportData.NumLibsPassNoBoards = reportData.NumLibsPassNoBoards + 1
		}
		if totFailClaim > 0 {
			reportData.NumLibsFailClaim = reportData.NumLibsFailClaim + 1
		}
		if totFail > 0 && compatibilityAsterisk[lib] {
			reportData.NumLibsAsteriskFail = reportData.NumLibsAsteriskFail + 1
		}
	}

	// Examples statistics
	for num, cnt := range numExamples {
		reportData.Examples = append(reportData.Examples, exampleReportData{num, cnt})
	}
	sort.Slice(reportData.Examples, func(i, j int) bool {
		return reportData.Examples[i].Num < reportData.Examples[j].Num
	})

	// Output results to console
	fmt.Printf("Tested libraries: %d\n\n", reportData.NumLibs)
	for _, c := range reportData.Boards {
		fmt.Printf("%s @ %s\n", c.Name, c.Versions)

		if c.Pass > 0 {
			fmt.Printf("- Compatible libs:          %d (%s)\n", c.Pass, percent(c.Pass))
			fmt.Printf("    claiming compatibility: %d (%s)\n", c.PassClaim, percent(c.PassClaim))
		}
		if c.Fail > 0 {
			fmt.Printf("- Incompatible libs:        %d (%s)\n", c.Fail, percent(c.Fail))
			fmt.Printf("    claiming compatibility: %d (%s)\n", c.FailClaim, percent(c.FailClaim))
		}
		if c.Untested > 0 {
			fmt.Printf("- Untested libs:            %d (%s)\n", c.Untested, percent(c.Untested))
		}
	}
	fmt.Printf("\n")
	fmt.Printf("Number of examples (distribution):\n")
	for _, e := range reportData.Examples {
		fmt.Printf("- %d: %d (%s)\n", e.Num, e.Count, percent(e.Count))
	}

	// Write an HTML report
	{
		templ, err := template.New("report").Funcs(template.FuncMap{"percent": percent}).Parse(htmlTmpl)
		if err != nil {
			panic(err)
		}
		os.Mkdir(outputDir, os.ModePerm)
		f, err := os.Create(path.Join(outputDir, "index.html"))
		defer f.Close()
		err = templ.Execute(f, reportData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating HTML report: %v\n", err)
			os.Exit(1)
		}
	}

	// Write per-library reports
	templ, err := template.New("report").Funcs(template.FuncMap{"percent": percent}).Parse(htmlTmplLibrary)
	if err != nil {
		panic(err)
	}
	for _, lib := range reportData.Libraries {
		f, err := os.Create(path.Join(outputDir, lib.ReportFile))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %s: %v\n", lib.ReportFile, err)
			os.Exit(1)
		}
		data := struct {
			Timestamp string
			Lib       libraryReportData
			Boards    []boardReportData
		}{
			Timestamp: reportData.Timestamp,
			Lib:       lib,
			Boards:    reportData.Boards,
		}
		err = templ.Execute(f, data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating HTML report: %s: %v\n", lib.ReportFile, err)
			os.Exit(1)
		}
		f.Close()
	}
	fmt.Printf("\nHTML report written to %s\n", path.Join(outputDir, "index.html"))
}

// This is the same function used to generate the library directory on the Arduino.cc website
func libraryURL(name string) string {
	name = strings.Replace(strings.TrimSpace(name), " ", "-", -1)
	name = strings.ToLower(name)
	name = regexp.MustCompile("[^a-z0-9_-]").ReplaceAllString(name, "")
	return "https://www.arduino.cc/reference/en/libraries/" + name + "/"
}
