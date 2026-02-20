// Package module provides internal types and functions for the GNyx application.
package module

import (
	"github.com/kubex-ecosystem/domus/cmd/cli"
	"github.com/kubex-ecosystem/domus/internal/module/version"

	info "github.com/kubex-ecosystem/domus/internal/module/info"
	kbxInfo "github.com/kubex-ecosystem/kbx/tools/info"
	kbxStyle "github.com/kubex-ecosystem/kbx/tools/style"
	logz "github.com/kubex-ecosystem/logz"

	"github.com/spf13/cobra"
)

type Domus struct {
	parentCmdName string
	hideBanner    bool
	certPath      string
	keyPath       string
	configPath    string
}

func (m *Domus) Alias() string {
	return ""
}
func (m *Domus) ShortDescription() string {
	return "KubexDomus: GKBX Database and Docker manager/service. "
}
func (m *Domus) LongDescription() string {
	return `KubexDomus: Is a tool to manage GKBX database and Docker services. It provides many DB flavors like MySQL, PostgreSQL, MongoDB, Redis, etc. It also provides Docker services like Docker Swarm, Docker Compose, etc. It is a command line tool that can be used to manage GKBX database and Docker services.`
}
func (m *Domus) Usage() string {
	return "domus [command] [args]"
}
func (m *Domus) Examples() []string {
	return []string{"domus [command] [args]", "domus database user auth'", "domus db roles list"}
}
func (m *Domus) Active() bool {
	return true
}
func (m *Domus) Module() string {
	return "domus"
}
func (m *Domus) Execute() error {
	dbChanData := make(chan interface{})
	defer close(dbChanData)

	if spyderErr := m.Command().Execute(); spyderErr != nil {
		logz.Log("error", spyderErr.Error())
		return spyderErr
	} else {
		return nil
	}
}
func (m *Domus) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use: m.Module(),
		//Aliases:     []string{m.Alias(), "w", "wb", "webServer", "http"},
		Example: m.concatenateExamples(),
		Annotations: kbxInfo.CLIBannerStyle(
			info.GetBanners(),
			[]string{
				m.LongDescription(),
				m.ShortDescription(),
			}, m.hideBanner,
		),
		Version: version.GetVersion(),
		Run: func(cmd *cobra.Command, args []string) {
			_ = cmd.Help()
		},
	}

	cmd.AddCommand(version.CliCommand())
	// cmd.AddCommand(cli.NewServiceCommand())
	cmd.AddCommand(cli.DockerCmd())
	cmd.AddCommand(cli.DatabaseCmd())
	cmd.AddCommand(cli.UtilsCmds())
	cmd.AddCommand(cli.SSHCmds())
	cmd.AddCommand(cli.ConfigCmd())

	kbxStyle.SetUsageTemplate(cmd)

	return cmd
}

func (m *Domus) SetParentCmdName(rtCmd string) {
	m.parentCmdName = rtCmd
}
func (m *Domus) concatenateExamples() string {
	examples := ""
	rtCmd := m.parentCmdName
	if rtCmd != "" {
		rtCmd = rtCmd + " "
	}
	for _, example := range m.Examples() {
		examples += rtCmd + example + "\n  "
	}
	return examples
}
