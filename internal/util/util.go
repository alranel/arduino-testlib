package util

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/alranel/arduino-testlib/internal/configuration"
	"github.com/arduino/arduino-cli/arduino/utils"
)

func LibrariesDirectory() string {
	return path.Join(configuration.CLIUserDir, "libraries")
}

func LibPathFromName(name string) string {
	return path.Join(LibrariesDirectory(), utils.SanitizeName(name))
}

func CoreFromFQBN(fqbn string) string {
	return strings.Join(strings.Split(fqbn, ":")[0:2], ":")
}

func ArchitectureFromFQBN(fqbn string) string {
	return strings.Split(fqbn, ":")[1]
}

// CoreInArchitectures returns true if the given core is compatible with the
// given list of architectures.
func CoreInArchitectures(core string, architectures []string) bool {
	coreArch := strings.SplitN(core, ":", 2)[1]
	for _, arch := range architectures {
		if arch == "*" || strings.ToLower(arch) == coreArch {
			return true
		}
	}
	return false
}

func runCLI(cliCmd []string) bool {
	fmt.Println("==> " + strings.Join(cliCmd, " "))
	cmd := exec.Command(cliCmd[0], cliCmd[1:]...)
	out, err := cmd.CombinedOutput()
	fmt.Printf("%s\n", out)
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			fmt.Println(err)
			os.Exit(1)
		}
	}
	return cmd.ProcessState.ExitCode() == 0
}
