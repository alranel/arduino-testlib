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
		numExamples[nEx] = numExamples[nEx] + 1
	}

	// Compute statistics
	numLibs := len(libraries)
	percent := func(n int) string { return fmt.Sprintf("%.1f%%", float32(n)/float32(numLibs)*100) }

	type boardReportData struct {
		Name                                             string
		Versions                                         string
		TotPass, PassClaim, TotFail, FailClaim, Untested int
		TotPassPercent, PassClaimPercent, PassNoClaimPercent, TotFailPercent,
		FailClaimPercent, TotClaimMismatchPercent, UntestedPercent string
	}
	type libraryReportData struct {
		Name, ReportFile, Version, URL string
		BoardCompatibility             map[string]compatibilityStatus
		BoardTestResults               map[string]test.TestResult
		Examples                       []string
	}
	type exampleReportData struct {
		Num, Count   int
		CountPercent string
	}
	reportData := struct {
		Timestamp                                       string
		NumLibs, NumBoards                              int
		NumLibsNoBoards, NumLibsAllBoards               int
		NumLibsNoBoardsPercent, NumLibsAllBoardsPercent string
		HasUntested                                     bool
		Boards                                          []boardReportData
		Examples                                        []exampleReportData
		Libraries                                       []libraryReportData
	}{
		Timestamp: time.Now().Format(time.RFC850),
		NumLibs:   numLibs,
		NumBoards: len(boards),
	}

	// Board statistics
	for board := range boards {
		var versions []string
		for v := range boards[board] {
			versions = append(versions, v)
		}
		semver.Sort(versions)

		c := boardReportData{
			Name:     board,
			Versions: strings.Join(versions, ","),
		}

		cnt := make(map[compatibilityStatus]int)
		for pair, result := range compatibility {
			if pair.board == board {
				cnt[result] = cnt[result] + 1
			}
		}
		c.TotPass = cnt[PASS_CLAIM] + cnt[PASS_NOCLAIM]
		c.TotPassPercent = percent(c.TotPass)
		c.PassClaim = cnt[PASS_CLAIM]
		c.PassClaimPercent = percent(cnt[PASS_CLAIM])
		c.PassNoClaimPercent = percent(cnt[PASS_NOCLAIM])
		c.TotClaimMismatchPercent = percent(cnt[PASS_NOCLAIM] + cnt[FAIL_CLAIM])
		c.TotFail = cnt[FAIL_CLAIM] + cnt[FAIL_NOCLAIM]
		c.TotFailPercent = percent(c.TotFail)
		c.FailClaim = cnt[FAIL_CLAIM]
		c.FailClaimPercent = percent(cnt[FAIL_CLAIM])
		c.Untested = numLibs - (c.TotPass + c.TotFail)
		c.UntestedPercent = percent(c.Untested)
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
		totPass := 0
		for pair, result := range compatibility {
			if pair.lib == lib {
				lData.BoardCompatibility[pair.board] = result
				if result == PASS_CLAIM || result == PASS_NOCLAIM {
					totPass = totPass + 1
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
		if totPass == len(boards) {
			reportData.NumLibsAllBoards = reportData.NumLibsAllBoards + 1
		}
		if totPass == 0 {
			reportData.NumLibsNoBoards = reportData.NumLibsNoBoards + 1
		}
	}
	reportData.NumLibsAllBoardsPercent = percent(reportData.NumLibsAllBoards)
	reportData.NumLibsNoBoardsPercent = percent(reportData.NumLibsNoBoards)

	// Examples statistics
	for num, cnt := range numExamples {
		reportData.Examples = append(reportData.Examples, exampleReportData{num, cnt, percent(cnt)})
	}
	sort.Slice(reportData.Examples, func(i, j int) bool {
		return reportData.Examples[i].Num < reportData.Examples[j].Num
	})

	// Output results to console
	fmt.Printf("Tested libraries: %d\n\n", reportData.NumLibs)
	for _, c := range reportData.Boards {
		fmt.Printf("%s @ %s\n", c.Name, c.Versions)

		if c.TotPass > 0 {
			fmt.Printf("- Compatible libs:          %d (%s)\n", c.TotPass, c.TotPassPercent)
			fmt.Printf("    claiming compatibility: %d (%s)\n", c.PassClaim, c.PassClaimPercent)
		}
		if c.TotFail > 0 {
			fmt.Printf("- Incompatible libs:        %d (%s)\n", c.TotFail, c.TotFailPercent)
			fmt.Printf("    claiming compatibility: %d (%s)\n", c.FailClaim, c.FailClaimPercent)
		}
		if c.Untested > 0 {
			fmt.Printf("- Untested libs:            %d (%s)\n", c.Untested, c.UntestedPercent)
		}
	}
	fmt.Printf("\n")
	fmt.Printf("Number of examples (distribution):\n")
	for _, e := range reportData.Examples {
		fmt.Printf("- %d: %d (%s)\n", e.Num, e.Count, e.CountPercent)
	}

	// Write an HTML report
	{
		templ, err := template.New("report").Parse(htmlTmpl)
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
	templ, err := template.New("report").Parse(htmlTmplLibrary)
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
