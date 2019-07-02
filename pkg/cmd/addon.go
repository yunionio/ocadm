package cmd

import (
	"github.com/pkg/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	"yunion.io/x/onecloud/pkg/mcclient"

	apiv1 "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	baremetaladdon "yunion.io/x/ocadm/pkg/phases/addons/baremetal"
)

// NewAddonPhase returns the addon Cobra command
func NewAddonPhase() workflow.Phase {
	return workflow.Phase{
		Name:  "addon",
		Short: "Installs required addons for kubernetes and onecloud cluster",
		Phases: []workflow.Phase{
			{
				Name:           "all",
				Short:          "Installs all the addons",
				InheritFlags:   getAddonPhaseFlags("all"),
				RunAllSiblings: true,
			},
			{
				Name:  "baremetal",
				Short: "Installs the OneCloud baremetal management service",
				Run:   runBaremetalAddon,
			},
		},
	}
}

func getInitData(c workflow.RunData) (*apiv1.InitConfiguration, clientset.Interface, *mcclient.ClientSession, error) {
	data, ok := c.(initData)
	if !ok {
		return nil, nil, nil, errors.New("addon phase invoked with an invalid data struct")
	}
	cfg := data.OnecloudCfg()
	client, err := data.Client()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "kubernetes client")
	}
	session, err := data.OnecloudClientSession()
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "onecloud session")
	}
	return cfg, client, session, nil
}

// runBaremetalAddon installs OneCloud addon to a Kubernetes cluster
func runBaremetalAddon(c workflow.RunData) error {
	cfg, client, session, err := getInitData(c)
	if err != nil {
		return err
	}
	return baremetaladdon.EnsureBaremetalAddon(cfg, client, session)
}

func getAddonPhaseFlags(name string) []string {
	flags := []string{
		options.CfgPath,
		options.KubeconfigPath,
		options.KubernetesVersion,
		options.ImageRepository,
	}
	// TODO: support all kinds of addons
	return flags
}
