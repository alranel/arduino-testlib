package test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
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

func TestLib(libName string, tr TestResults) TestResults {
	libPath := util.LibPathFromName(libName)
	if _, err := os.Stat(libPath); err != nil {
		fmt.Fprintln(os.Stderr, "Library not found")
		os.Exit(1)
	}

	// Get library name
	properties, err := ini.Load(path.Join(libPath, "library.properties"))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	name := properties.Section("").Key("name").String()
	version := properties.Section("").Key("version").String()
	architectures := strings.Split(properties.Section("").Key("architectures").String(), ",")
	includes := strings.Split(properties.Section("").Key("includes").String(), ",")
	if name == "" {
		fmt.Println("No library name found in library.properties")
		os.Exit(1)
	}

	if tr.Name != "" && tr.Name != name {
		fmt.Fprintf(os.Stderr, "Library name mismatch; known: %s, tested: %s\n", tr.Name, name)
		os.Exit(1)
	}
	tr.Name = name

	// Look for a main header file
	headerFile := utils.SanitizeName(name) + ".h"
	headerFilePath := path.Join(libPath, "src", headerFile)
	headerFileCreated := false
	if _, err := os.Stat(headerFilePath); err != nil {
		// Check if the header file is in the root directory (old library format),
		// otherwise just create an empty header file. This will still allow the
		// compilation of .cpp files
		if _, err := os.Stat(path.Join(libPath, headerFile)); err != nil {
			fmt.Printf("Main header file not found, creating an empty one: %s\n", headerFile)
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
		coreVersion, err := cliclient.GetInstalledCoreVersion(core)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get core version for %s: %v\n", core, err)
			os.Exit(1)
		}
		fmt.Printf("Testing %s @ %s on %s @ %s\n", name, version, fqbn, coreVersion)

		// Check if this combo was already tested
		for _, t := range tr.Tests {
			if t.Version == version && t.FQBN == fqbn && t.CoreVersion == coreVersion {
				fmt.Printf("Skipping, combination already tested\n")
				continue fqbn
			}
		}

		// Test library inclusion
		resB, out := cliclient.CompileSketch(sketchDir, libPath, fqbn)
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
		examples, err := ioutil.ReadDir(path.Join(libPath, "examples"))
		if err == nil {
			for _, example := range examples {
				if !example.IsDir() {
					continue
				}
				resB, out := cliclient.CompileSketch(path.Join(libPath, "examples", example.Name()), libPath, fqbn)
				var res CompilationResult
				if resB {
					res = PASS
				} else {
					res = FAIL
				}
				result.Examples = append(result.Examples, exampleResult{
					Name:   example.Name(),
					Result: res,
					Log:    out,
				})
			}
		}

		tr.Tests = append(tr.Tests, result)
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
