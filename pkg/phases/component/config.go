package component

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

func NewCmdConfig(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage components config",
	}
	NewConfigCmd(cmd, out).Bind()
	return cmd
}

type ConfigCmd struct {
	*baseCmd
}

func NewConfigCmd(cmd *cobra.Command, out io.Writer) *ConfigCmd {
	return &ConfigCmd{
		baseCmd: newBaseCmd(cmd, out),
	}
}

func (c ConfigCmd) Bind() {
	c.baseCmd.AddCmd(c.newSubCmd("show", c.show))
	c.baseCmd.AddCmd(c.newSubCmd("edit", c.edit))
}

func (c ConfigCmd) getComponentsConfig() *OnecloudComponentsConfig {
	return c.data.cfg
}

func (c ConfigCmd) getComponentsConfigString() (string, error) {
	return c.getComponentsConfig().ToYaml()
}

func (c ConfigCmd) show(_ *componentsData, out io.Writer) error {
	data, err := c.getComponentsConfigString()
	if err != nil {
		return err
	}
	fmt.Fprintf(out, "%s", data)
	return nil
}

func (c ConfigCmd) edit(_ *componentsData, _ io.Writer) error {
	data, err := c.getComponentsConfigString()
	if err != nil {
		return err
	}
	tempfile, err := ioutil.TempFile("", "components-config.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tempfile.Name())
	if _, err := tempfile.Write([]byte(data)); err != nil {
		return err
	}
	cmd := exec.Command("vim", tempfile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		return err
	}
	content, err := ioutil.ReadFile(tempfile.Name())
	if err != nil {
		return err
	}
	cfg, err := NewOnecloudComponentsConfigFromYaml(string(content))
	if err != nil {
		return err
	}
	oc := c.data.OnecloudCluster()
	cfgMap, err := cfg.ToConfigMap(oc)
	if err != nil {
		return err
	}
	return SyncConfigMap(c.data.KubernetesClient(), oc, cfgMap)
}
