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
log_level={{ .General.LogLevel }}
password_hash_iterations={{ .General.PasswordHashIterations }}

[postgresql]
dsn="{{ .PostgreSQL.DSN }}"
automigrate={{ .PostgreSQL.Automigrate }}

[redis]
url="{{ .Redis.URL }}"
max_idle={{ .Redis.MaxIdle }}
idle_timeout="{{ .Redis.IdleTimeout }}"

[tdma_server]
bind="{{ .TdmaServer.Bind }}"
  [network_server.scheduler]
  scheduler_interval="{{ .TdmaServer.Scheduler.SchedulerInterval }}"

[app_server]
bind="{{ .AppServer.Bind }}"
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
