package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/alranel/arduino-testlib/internal/configuration"
	"github.com/arduino/arduino-cli/arduino/utils"
	"github.com/spf13/cobra"
	semver "go.bug.st/relaxed-semver"
	"gopkg.in/ini.v1"
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

	/*
		// This implementation based on arduino-cli works but it's quite slow because
		// arduino-cli is not optimized for bulk install (it reloads everything after
		// the installation of each library).
		instance := cliclient.NewInstance()
		for _, lib := range instance.GetAllLibraries() {
			// Use the unsanitized name to install the library
			instance.InstallLibrary(lib, "")
		}
	*/

	// Read the library index
	type library struct {
		Name, Version, URL, ArchiveFileName string
	}
	libraries := make(map[string]library)
	var jsonData struct{ Libraries []library }

	jsonFile, err := os.Open(path.Join(configuration.CLIDataDir, "data/library_index.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to read library_index.json: %v\n", err)
		os.Exit(1)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &jsonData)

	// Find last version for each library
	for _, lib := range jsonData.Libraries {
		if l, ok := libraries[lib.Name]; ok {
			if semver.ParseRelaxed(lib.Version).GreaterThan(semver.ParseRelaxed(l.Version)) {
				libraries[lib.Name] = lib
			}
		} else {
			libraries[lib.Name] = lib
		}
	}

	// Download and unzip libraries
	os.MkdirAll(path.Join(configuration.CLIDataDir, "downloads/libraries"), os.ModePerm)
	os.MkdirAll(path.Join(configuration.CLIUserDir, "libraries"), os.ModePerm)
	for libName, lib := range libraries {
		// Check if we already have this version and skip download
		libPath := path.Join(configuration.CLIUserDir, "libraries", utils.SanitizeName(libName))
		if _, err := os.Stat(libPath); !os.IsNotExist(err) {
			properties, err := ini.Load(path.Join(libPath, "library.properties"))
			if err != nil {
				if properties.Section("").Key("version").String() == lib.Version {
					fmt.Printf("Skipping %s@%s\n", libName, lib.Version)
					continue
				}
			}
		}

		fmt.Printf("Installing %s@%s\n", libName, lib.Version)

		resp, err := http.Get(lib.URL)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()
		filename := path.Join(configuration.CLIDataDir, "downloads/libraries", lib.ArchiveFileName)
		out, _ := os.Create(filename)
		defer out.Close()
		io.Copy(out, resp.Body)

		_, err = unzip(filename, libPath)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func unzip(src string, destination string) ([]string, error) {
	os.MkdirAll(destination, os.ModePerm)

	var filenames []string
	r, err := zip.OpenReader(src)

	if err != nil {
		return filenames, err
	}

	defer r.Close()

	for _, f := range r.File {
		// Remove root directory such as Foo-1.0.0 so that we put everything
		// in the destination folder
		f.Name = strings.SplitN(f.Name, "/", 2)[1]
		if f.Name == "" {
			continue
		}
		fpath := filepath.Join(destination, f.Name)

		if !strings.HasPrefix(fpath, filepath.Clean(destination)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s is an illegal filepath (%s)", fpath, filepath.Clean(destination)+string(os.PathSeparator))
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath,
			os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
			f.Mode())

		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()

		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}

	return filenames, nil
}
