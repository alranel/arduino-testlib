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
)

var instance *cli_rpc.Instance

func Init() {
	cli_conf.Settings = cli_conf.Init("")
	cli_conf.Settings.Set("logging.level", "error")
	cli_conf.Settings.Set("directories.Data", path.Join(configuration.CLIDataDir, "data"))
	cli_conf.Settings.Set("directories.Downloads", path.Join(configuration.CLIDataDir, "downloads"))
	cli_conf.Settings.Set("directories.User", path.Join(configuration.CLIDataDir, "user"))
	if configuration.AdditionalURLs != "" {
		cli_conf.Settings.Set("board_manager.additional_urls", strings.Split(configuration.AdditionalURLs, ","))
	}
	instance = cli_instance.CreateAndInit()

	// Update index
	{
		_, err := cli_commands.UpdateIndex(context.Background(), &cli_rpc.UpdateIndexRequest{
			Instance: instance,
		}, cli_output.ProgressBar())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating index: %v", err)
			os.Exit(1)
		}
	}

	// Update library index
	{
		err := cli_commands.UpdateLibrariesIndex(context.Background(), &cli_rpc.UpdateLibrariesIndexRequest{
			Instance: instance,
		}, cli_output.ProgressBar())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error updating library index: %v", err)
			os.Exit(1)
		}
	}
}

func InstallLibrary(libName string, version string) bool {
	fmt.Printf("=> Installing lib: %s\n", libName)

	libraryInstallRequest := &cli_rpc.LibraryInstallRequest{
		Instance: instance,
		Name:     libName,
		Version:  version,
		NoDeps:   true,
	}
	err := cli_lib.LibraryInstall(context.Background(), libraryInstallRequest, cli_output.ProgressBar(), cli_output.TaskProgress())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error installing %s: %v", libName, err)
		return false
	}

	return true
}

func InstallCores() {
	for _, fqbn := range configuration.FQBNs {
		t := strings.Split(fqbn, ":")
		platformInstallRequest := &cli_rpc.PlatformInstallRequest{
			Instance:        instance,
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

func GetAllLibraries() []string {
	res, err := cli_lib.LibraryList(context.Background(), &cli_rpc.LibraryListRequest{
		Instance: instance,
		All:      true,
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

func GetInstalledCoreVersion(core string) (string, error) {
	platforms, err := cli_core.GetPlatforms(&cli_rpc.PlatformListRequest{
		Instance:      instance,
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

func CompileSketch(sketchPath string, libPath string, fqbn string) (result bool, out string) {
	compileRequest := &cli_rpc.CompileRequest{
		Instance:   instance,
		Fqbn:       fqbn,
		SketchPath: sketchPath,
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
