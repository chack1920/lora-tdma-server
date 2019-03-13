package cmd

import (
	"os"
	"text/template"

	"github.com/lioneie/lora-tdma-server/internal/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// when updating this template, don't forget to update config.md!
const configTemplate = `[general]
# Log level
#
# debug=5, info=4, warning=3, error=2, fatal=1, panic=0
log_level={{ .General.LogLevel }}

# The number of times passwords must be hashed. A higher number is safer as
# an attack takes more time to perform.
password_hash_iterations={{ .General.PasswordHashIterations }}

# Tdma-server configuration.
[tdma_server]
# ip:port to bind the tdma-server api interface to
bind="{{ .TdmaServer.Bind }}"
`

var configCmd = &cobra.Command{
	Use:   "configfile",
	Short: "Print the LoRa TDMA Server configuration file",
	RunE: func(cmd *cobra.Command, args []string) error {
		t := template.Must(template.New("config").Parse(configTemplate))
		err := t.Execute(os.Stdout, config.C)
		if err != nil {
			return errors.Wrap(err, "execute config template error")
		}
		return nil
	},
}
