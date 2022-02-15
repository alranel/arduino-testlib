package configuration

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/pflag"
	"gopkg.in/yaml.v1"
)

var CLIDataDir, CLIUserDir string
var AdditionalURLs string
var FQBNs []string

func Initialize(flags *pflag.FlagSet) error {
	CLIDataDir, _ = flags.GetString("cli-datadir")

	// If no custom CLI datadir was specified, read the default paths from arduino-cli
	if CLIDataDir == "" {
		cliCmd := []string{"arduino-cli", "config", "dump"}
		cmd := exec.Command(cliCmd[0], cliCmd[1:]...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		cliConfig := make(map[string]map[string]string)
		err = yaml.Unmarshal(out, &cliConfig)
		CLIUserDir = cliConfig["directories"]["user"]
	} else {
		CLIUserDir = path.Join(CLIDataDir, "user")
		os.Setenv("ARDUINO_DIRECTORIES_DATA", path.Join(CLIDataDir, "data"))
		os.Setenv("ARDUINO_DIRECTORIES_DOWNLOADS", path.Join(CLIDataDir, "downloads"))
		os.Setenv("ARDUINO_DIRECTORIES_USER", CLIUserDir)
	}
	//fmt.Printf("CLI user dir = %s\n", CLIUserDir)

	AdditionalURLs, _ = flags.GetString("additional-urls")
	FQBNs, _ = flags.GetStringSlice("fqbn")

	return nil
}
