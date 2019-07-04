package join

import (
	"github.com/pkg/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"
	kubeconfigutil "k8s.io/kubernetes/cmd/kubeadm/app/util/kubeconfig"
	"yunion.io/x/ocadm/pkg/phases/uploadconfig"

	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/copycerts"
)

func NewNodePreparePhase() workflow.Phase {
	return workflow.Phase{
		Name:  "node-join-prepare",
		Short: "Prepares the machine for join.",
		Phases: []workflow.Phase{
			{
				Name:           "all [api-server-endpoint]",
				Short:          "Prepares the machine for serving as a join node.",
				InheritFlags:   getJoinNodePreparePhaseFlags("all"),
				RunAllSiblings: true,
			},
			newJoinNodePrepareDownloadCertsSubphase(),
			newJoinNodePrepareDownloadConfigSubPhase(),
		},
	}
}

func getJoinNodePreparePhaseFlags(name string) []string {
	var flags []string
	switch name {
	case "all":
		flags = []string{
			options.APIServerAdvertiseAddress,
			options.APIServerBindPort,
			options.CfgPath,
			options.FileDiscovery,
			options.TokenDiscovery,
			options.TokenDiscoveryCAHash,
			options.TokenDiscoverySkipCAHash,
			options.TLSBootstrapToken,
			options.TokenStr,
			options.CertificateKey,
		}
	case "download-certs", "download-config":
		flags = []string{
			options.CfgPath,
			options.ControlPlane,
			options.FileDiscovery,
			options.TokenDiscovery,
			options.TokenDiscoveryCAHash,
			options.TokenDiscoverySkipCAHash,
			options.TLSBootstrapToken,
			options.TokenStr,
			options.CertificateKey,
		}
	default:
		flags = []string{}
	}
	return flags
}

func newJoinNodePrepareDownloadCertsSubphase() workflow.Phase {
	return workflow.Phase{
		Name:         "download-certs [apiserver-endpoint]",
		Run:          runJoinNodePrepareDownloadCertsPhaseLocal,
		InheritFlags: getJoinNodePreparePhaseFlags("download-certs"),
	}
}

func runJoinNodePrepareDownloadCertsPhaseLocal(c workflow.RunData) error {
	data, ok := c.(JoinData)
	if !ok {
		return errors.New("download-certs phase invoked with an invalid data struct")
	}
	cfg, err := data.OnecloudInitCfg()
	if err != nil {
		return err
	}

	client, err := bootstrapClient(data)
	if err != nil {
		return err
	}

	if err := copycerts.DownloadCerts(client, cfg); err != nil {
		return errors.Wrap(err, "error downloading certs")
	}
	return nil
}

func newJoinNodePrepareDownloadConfigSubPhase() workflow.Phase {
	return workflow.Phase{
		Name:         "download-config [apiserver-endpoint]",
		Run:          runJoinNodePrepareDownloadRCAdminPhaseLocal,
		InheritFlags: getJoinNodePreparePhaseFlags("download-config"),
	}
}

func runJoinNodePrepareDownloadRCAdminPhaseLocal(c workflow.RunData) error {
	data, ok := c.(JoinData)
	if !ok {
		return errors.New("download-rcadm phase invoked with an invalid data struct")
	}

	client, err := bootstrapClient(data)
	if err != nil {
		return err
	}
	return uploadconfig.DownloadConfiguration(client)
}

func bootstrapClient(data JoinData) (clientset.Interface, error) {
	tlsBootstrapCfg, err := data.TLSBootstrapCfg()
	if err != nil {
		return nil, errors.Wrap(err, "unable to access the cluster")
	}
	client, err := kubeconfigutil.ToClientSet(tlsBootstrapCfg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to access the cluster")
	}
	return client, nil
}
