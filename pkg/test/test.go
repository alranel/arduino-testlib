package test

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alranel/arduino-testlib/internal/cliclient"
	"github.com/alranel/arduino-testlib/internal/configuration"
	"github.com/alranel/arduino-testlib/internal/util"
	"github.com/arduino/arduino-cli/arduino/utils"
	"gopkg.in/ini.v1"
)

type CompilationResult string

const (
	PASS CompilationResult = "PASS"
	FAIL CompilationResult = "FAIL"
)

type exampleResult struct {
	Name   string            `json:"name"`
	Result CompilationResult `json:"result"`
	Log    string            `json:"log"`
}

type TestResult struct {
	Version       string            `json:"version"`
	Architectures []string          `json:"architectures"`
	FQBN          string            `json:"fqbn"`
	Core          string            `json:"core"`
	CoreVersion   string            `json:"core_version"`
	Result        CompilationResult `json:"result"`
	Log           string            `json:"log"`
	Examples      []exampleResult   `json:"examples"`
	NoMainHeader  bool              `json:"no_main_header"`
}

type TestResults struct {
	Name  string       `json:"name"`
	Tests []TestResult `json:"tests"`
}

func TestLibByName(libName string, tr TestResults, force bool, instance *cliclient.CliInstance) TestResults {
	libPath := util.LibPathFromName(libName)
	return TestLib(libPath, tr, force, instance)
}

func TestLib(libPath string, tr TestResults, force bool, instance *cliclient.CliInstance) TestResults {
	libPath, _ = filepath.Abs(libPath)
	if _, err := os.Stat(libPath); err != nil {
		fmt.Fprintf(os.Stderr, "Library not found in directory: %s\n", libPath)
		return tr
	}

	// Get library name
	properties, err := ini.Load(path.Join(libPath, "library.properties"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open library.properties: %s\n", libPath)
		return tr
	}
	name := properties.Section("").Key("name").String()
	version := properties.Section("").Key("version").String()
	nameAndVersion := name + "@" + version
	architectures := strings.Split(properties.Section("").Key("architectures").String(), ",")
	includes := strings.Split(properties.Section("").Key("includes").String(), ",")
	if name == "" {
		fmt.Printf("No library name found in library.properties: %s\n", libPath)
		return tr
	}

	if tr.Name != "" && tr.Name != name {
		fmt.Fprintf(os.Stderr, "[%s] Library name mismatch; known: %s, tested: %s\n", nameAndVersion, tr.Name, name)
		os.Exit(1)
	}
	tr.Name = name
	fmt.Printf("[%s] Start testing\n", nameAndVersion)

	// Look for a main header file
	headerFile := utils.SanitizeName(name) + ".h"
	headerFilePath := path.Join(libPath, "src", headerFile)
	headerFileCreated := false
	if _, err := os.Stat(headerFilePath); err != nil {
		// Check if the header file is in the root directory (old library format),
		// otherwise just create an empty header file. This will still allow the
		// compilation of .cpp files
		if _, err := os.Stat(path.Join(libPath, headerFile)); err != nil {
			fmt.Printf("[%s] Main header file not found, creating an empty one: %s\n", nameAndVersion, headerFile)
			os.Create(headerFilePath)
			headerFileCreated = true
		}
	}

	// Create a bogus sketch in a temporary directory
	tmpDir, err := ioutil.TempDir("/tmp", "arduino-testlib")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)
	sketchDir := path.Join(tmpDir, "test")
	os.Mkdir(sketchDir, os.ModePerm)

	// Generate the sketch content
	var sketch string
	if len(includes) > 0 && includes[0] != "" {
		for _, include := range includes {
			if include == "" {
				continue
			}
			sketch = sketch + "#include <" + strings.TrimSpace(include) + ">\n"
		}
	} else {
		sketch = sketch + "#include <" + headerFile + ">\n"
	}
	sketch = sketch + "void setup() {}\n"
	sketch = sketch + "void loop() {}\n"
	f, _ := os.Create(path.Join(sketchDir, "test.ino"))
	f.WriteString(sketch)
	f.Close()

	// Try to compile the sketch
fqbn:
	for _, fqbn := range configuration.FQBNs {
		core := util.CoreFromFQBN(fqbn)
		coreVersion, err := instance.GetInstalledCoreVersion(core)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[%s] Failed to get core version for %s: %v\n", nameAndVersion, core, err)
			os.Exit(1)
		}

		// Check if this combo was already tested
		if !force {
			for _, t := range tr.Tests {
				if t.Version == version && t.FQBN == fqbn && t.CoreVersion == coreVersion {
					fmt.Printf("[%s] skipping %s, already tested\n", nameAndVersion, fqbn)
					continue fqbn
				}
			}
		} else {
			// Remove past test results for this combo
			var tt []TestResult
			for _, t := range tr.Tests {
				if t.Version != version || t.FQBN != fqbn || t.CoreVersion != coreVersion {
					tt = append(tt, t)
				}
			}
			tr.Tests = tt
		}

		// Test library inclusion
		resB, out := instance.CompileSketch(sketchDir, libPath, fqbn)
		var res CompilationResult
		if resB {
			res = PASS
		} else {
			res = FAIL
		}

		// Store results
		result := TestResult{
			Version:       version,
			Architectures: architectures,
			FQBN:          fqbn,
			Core:          core,
			CoreVersion:   coreVersion,
			Log:           out,
			Result:        res,
			Examples:      []exampleResult{},
			NoMainHeader:  headerFileCreated,
		}

		// Test examples
		filepath.Walk(path.Join(libPath, "examples"), func(path string, info fs.FileInfo, err error) error {
			if strings.HasSuffix(info.Name(), ".ino") {
				exampleDir := filepath.Dir(path)
				resB, out := instance.CompileSketch(exampleDir, libPath, fqbn)
				var res CompilationResult
				if resB {
					res = PASS
				} else {
					res = FAIL
				}
				result.Examples = append(result.Examples, exampleResult{
					Name:   filepath.Base(exampleDir),
					Result: res,
					Log:    out,
				})
			}
			return nil
		})

		tr.Tests = append(tr.Tests, result)
	}

	{
		var results []string
		for _, t := range tr.Tests {
			results = append(results, t.FQBN+"="+string(t.Result))
		}
		fmt.Printf("[%s] %s\n", nameAndVersion, strings.Join(results, " "))
	}

	return tr
}

func ReadResultsFile(path string, tr *TestResults) bool {
	jsonFile, err := os.Open(path)
	if err != nil {
		return false
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &tr)
	return true
}
