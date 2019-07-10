package init

import (
	"github.com/spf13/cobra"

	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/ocadm/pkg/phases/components"
	"yunion.io/x/ocadm/pkg/phases/components/baremetal"
	"yunion.io/x/ocadm/pkg/phases/components/glance"
	"yunion.io/x/ocadm/pkg/phases/components/influxdb"
	"yunion.io/x/ocadm/pkg/phases/components/webconsole"
)

func init() {
	registerComponentCmds()
}

func registerComponentCmds() {
	installComponents := []*components.Component{
		glance.GlanceComponent,
		baremetal.BaremetalComponent,
		webconsole.WebconsoleComponent,
		influxdb.InfluxdbComponent,
	}
	for _, c := range installComponents {
		addCmdSubCmd(components.InstallCmd, c.ToInstallCmd(), c.ToInstallPhase())
	}
	components.InstallCmd.CompleteAllSubCmd()

	uninstallComponents := []*components.Component{
		glance.GlanceComponent,
		baremetal.BaremetalComponent,
		webconsole.WebconsoleComponent,
		influxdb.InfluxdbComponent,
	}
	for _, c := range uninstallComponents {
		addCmdSubCmd(components.UninstallCmd, c.ToUninstallCmd(), c.ToUninstallPhase())
	}
	components.UninstallCmd.CompleteAllSubCmd()
}

func addCmdSubCmd(actionCmd *components.ComponentActionCmd, cmd *cobra.Command, phase workflow.Phase) {
	actionCmd.AddCmd(&components.SubCmd{
		Cmd:   cmd,
		Phase: phase,
	})
}
