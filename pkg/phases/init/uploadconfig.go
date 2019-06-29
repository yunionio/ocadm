package init

import (
	"github.com/pkg/errors"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/kubeadm/app/cmd/phases/workflow"

	apis "yunion.io/x/ocadm/pkg/apis/v1"
	"yunion.io/x/ocadm/pkg/options"
	"yunion.io/x/ocadm/pkg/phases/uploadconfig"
)

// NewUploadConfigPhase returns the phase to upload onecloud config
func NewUploadConfigPhase() workflow.Phase {
	return workflow.Phase{
		Name:    "oc-upload-config",
		Aliases: []string{"ocuploadconfig"},
		Phases: []workflow.Phase{
			{
				Name:           "all",
				Short:          "Uploads all configuration to a config map",
				RunAllSiblings: true,
				InheritFlags:   getUploadConfigPhaseFlags(),
			},
			{
				Name:         "ocadm",
				Short:        "Uploads the ocadm InitConfiguration to a ConfigMap",
				Run:          runUploadOcadmConfig,
				InheritFlags: getUploadConfigPhaseFlags(),
			},
		},
	}
}

func getUploadConfigPhaseFlags() []string {
	return []string{
		options.CfgPath,
		options.KubeconfigPath,
	}
}

// runUploadOcadmConfig uploads the ocadm configuration to a ConfigMap
func runUploadOcadmConfig(c workflow.RunData) error {
	cfg, client, err := getUploadConfigData(c)
	if err != nil {
		return err
	}

	klog.V(1).Infoln("[oc-upload-config] Uploading the ocadm InitConfiguration to a ConfigMap")
	if err := uploadconfig.UploadConfiguration(cfg, client); err != nil {
		return errors.Wrap(err, "error uploading the ocadm InitConfiguration")
	}
	return nil
}

func getUploadConfigData(c workflow.RunData) (*apis.InitConfiguration, clientset.Interface, error) {
	data, ok := c.(InitData)
	if !ok {
		return nil, nil, errors.New("oc-upload-config phase invoked with an invalid data struct")
	}
	cfg := data.OnecloudCfg()
	client, err := data.Client()
	if err != nil {
		return nil, nil, err
	}
	return cfg, client, err
}
