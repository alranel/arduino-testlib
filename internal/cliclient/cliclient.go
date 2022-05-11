package cliclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/alranel/arduino-testlib/internal/configuration"
	cli_instance "github.com/arduino/arduino-cli/cli/instance"
	cli_output "github.com/arduino/arduino-cli/cli/output"
	cli_commands "github.com/arduino/arduino-cli/commands"
	cli_compile "github.com/arduino/arduino-cli/commands/compile"
	cli_core "github.com/arduino/arduino-cli/commands/core"
	cli_lib "github.com/arduino/arduino-cli/commands/lib"
	cli_conf "github.com/arduino/arduino-cli/configuration"
	cli_rpc "github.com/arduino/arduino-cli/rpc/cc/arduino/cli/commands/v1"
	"github.com/sirupsen/logrus"
)

type CliInstance struct {
	Instance *cli_rpc.Instance
}

func NewInstance() *CliInstance {
	cli_conf.Settings = cli_conf.Init("")
	logrus.SetLevel(logrus.ErrorLevel)
	cli_conf.Settings.Set("directories.Data", path.Join(configuration.CLIDataDir, "data"))
	cli_conf.Settings.Set("directories.Downloads", path.Join(configuration.CLIDataDir, "downloads"))
	cli_conf.Settings.Set("directories.User", path.Join(configuration.CLIDataDir, "user"))
	if configuration.AdditionalURLs != "" {
		cli_conf.Settings.Set("board_manager.additional_urls", strings.Split(configuration.AdditionalURLs, ","))
	}
	instance := new(CliInstance)
	instance.Instance = cli_instance.CreateAndInit()

	// Update index
	{
		_, err := cli_commands.UpdateIndex(context.Background(), &cli_rpc.UpdateIndexRequest{
			Instance: instance.Instance,
		}, cli_output.ProgressBar())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating index: %v", err)
			os.Exit(1)
		}
	}

	// Update library index
	{
		err := cli_commands.UpdateLibrariesIndex(context.Background(), &cli_rpc.UpdateLibrariesIndexRequest{
			Instance: instance.Instance,
		}, cli_output.ProgressBar())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating library index: %v", err)
			os.Exit(1)
		}
	}

	return instance
}

func (instance *CliInstance) InstallLibrary(libName string, version string) bool {
	fmt.Printf("=> Installing lib: %s\n", libName)

	libraryInstallRequest := &cli_rpc.LibraryInstallRequest{
		Instance: instance.Instance,
		Name:     libName,
		Version:  version,
		NoDeps:   true,
	}
	err := cli_lib.LibraryInstall(context.Background(), libraryInstallRequest, cli_output.ProgressBar(), cli_output.TaskProgress())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error installing %s: %v\n", libName, err)
		return false
	}

	return true
}

func (instance *CliInstance) InstallCores() {
	for _, fqbn := range configuration.FQBNs {
		t := strings.Split(fqbn, ":")
		platformInstallRequest := &cli_rpc.PlatformInstallRequest{
			Instance:        instance.Instance,
			PlatformPackage: t[0],
			Architecture:    t[1],
			Version:         "",
			SkipPostInstall: false,
		}
		_, err := cli_core.PlatformInstall(context.Background(), platformInstallRequest, cli_output.ProgressBar(), cli_output.TaskProgress())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error installing %s: %v", fqbn, err)
		}
	}
}

func (instance *CliInstance) GetAllLibraries() []string {
	res, err := cli_lib.LibrarySearch(context.Background(), &cli_rpc.LibrarySearchRequest{
		Instance: instance.Instance,
		Query:    "",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing libraries: %v", err)
		os.Exit(1)
	}

	var libs []string
	for _, lib := range res.GetLibraries() {
		libs = append(libs, lib.Name)
	}
	return libs
}

func (instance *CliInstance) GetInstalledLibraries() []string {
	res, err := cli_lib.LibraryList(context.Background(), &cli_rpc.LibraryListRequest{
		Instance: instance.Instance,
		All:      false,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing libraries: %v", err)
		os.Exit(1)
	}

	var libs []string
	for _, lib := range res.GetInstalledLibraries() {
		libs = append(libs, lib.Library.Name)
	}
	return libs
}

func (instance *CliInstance) GetInstalledCoreVersion(core string) (string, error) {
	platforms, err := cli_core.GetPlatforms(&cli_rpc.PlatformListRequest{
		Instance:      instance.Instance,
		UpdatableOnly: false,
		All:           true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing platforms: %v", err)
		os.Exit(1)
	}
	for _, p := range platforms {
		if p.Id == core {
			return p.Installed, nil
		}
	}
	return "", errors.New("Platform not found")
}

func (instance *CliInstance) CompileSketch(sketchPath string, libPath string, fqbn string) (result bool, out string) {
	compileRequest := &cli_rpc.CompileRequest{
		Instance:   instance.Instance,
		Fqbn:       fqbn,
		SketchPath: sketchPath,
		Library:    []string{libPath},
	}
	compileStdOut := new(bytes.Buffer)
	compileStdErr := new(bytes.Buffer)
	verboseCompile := false
	_, compileError := cli_compile.Compile(context.Background(), compileRequest, compileStdOut, compileStdErr, nil, verboseCompile)

	if compileError == nil {
		return true, compileStdOut.String() + compileStdErr.String()
	} else {
		return false, compileStdOut.String() + compileStdErr.String()
	}
}
